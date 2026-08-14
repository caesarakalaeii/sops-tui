{
  # Keep this line accurate and one line long: `nix flake metadata` prints it,
  # and it is the first thing a cold agent learns about the repo.
  description = "sops-tui -- k9s-inspired Go/Bubble Tea TUI for SOPS-encrypted secrets. Run `nix flake show` for the command map.";

  # nixpkgs is the only input, on purpose.
  #
  # flake-utils would buy exactly one thing here -- eachDefaultSystem -- which is
  # the three-line genAttrs below. In exchange it costs a second lock node in
  # every repo (flake-utils transitively pulls `systems`, so really two), a
  # second upstream that can break one repo and not the others, and a hardcoded
  # system list this repo cannot edit. That list is currently broken: it still
  # contains x86_64-darwin, which now throws (see `systems` below).
  #
  # nixos-unstable is the channel the author's own NixOS config tracks, so
  # `nix develop` here and `nixos-rebuild` there resolve the same store paths and
  # share one cache.
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    # `...` rather than a closed { self, nixpkgs }: adding a second input later
    # would otherwise fail with "called with unexpected argument 'self'".
    { nixpkgs, ... }:
    let
      lib = nixpkgs.lib;

      # x86_64-darwin is deliberately absent. nixpkgs 26.11 replaced that whole
      # attribute set with `throw "Nixpkgs 26.11 has dropped support for
      # x86_64-darwin"`. genAttrs is lazy, so plain `nix develop` on Linux would
      # not notice -- it detonates later, on `nix flake check --all-systems`.
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "aarch64-darwin"
      ];

      # Stand-in for flake-utils.lib.eachDefaultSystem. Passes `pkgs` rather than
      # a system string, because that is what every call site below wants.
      forAllSystems = f: lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});

      # ======================================================================
      # PER-REPO BLOCK 1 -- the toolchain
      # ======================================================================
      # Explicit `pkgs.foo`, never `with pkgs; [ ... ]`: when an attr disappears
      # in a nixpkgs bump, `with` reports a bare undefined identifier with no
      # hint of which set it came from, and the name is not greppable.
      #
      # `go` is the documented exception to pinning runtimes by major: there is
      # no major-pinned Go alias in nixpkgs (`go_1_24` does not exist), so bare
      # `pkgs.go` it is, kept honest by GOTOOLCHAIN=local below. go.mod asks for
      # go 1.26.2 and this pin ships 1.26.5, so the directive is satisfied
      # without any toolchain download.
      toolchain =
        pkgs:
        [
          # ---- Go ----
          pkgs.go
          pkgs.gopls
          pkgs.golangci-lint
          pkgs.delve

          # ---- what the program actually shells out to ----
          # internal/sops/executor.go and internal/app/model.go run `sops
          # decrypt|set|rotate|encrypt` via exec.CommandContext, and
          # internal/validator/startup.go hard-fails at startup when `sops` is
          # not on PATH. `age` is here for age-keygen: internal/sops/
          # setkey_e2e_test.go calls exec.LookPath("age-keygen") and SKIPS
          # itself when it is missing, so without this package `dev-test` would
          # report a green run that silently never exercised the sops round
          # trip. go-git is pure Go, but `git` is needed for $REPO_ROOT anyway.
          pkgs.sops
          pkgs.age

          # ---- present in every repo in the fleet ----
          pkgs.git
          pkgs.jq
          pkgs.gnumake
        ]
        # atotto/clipboard has no CGo and no library dependency -- it shells out
        # to a helper binary chosen in its init(): wl-copy/wl-paste when
        # WAYLAND_DISPLAY is set, else xclip, else xsel, and if it finds none it
        # sets clipboard.Unsupported, which internal/app/model.go checks before
        # the ctrl+y copy path. So both are needed to cover both session types,
        # and without them the TUI still runs -- that one keybinding just
        # degrades to an error message. Linux only: on darwin the same library
        # uses pbcopy/pbpaste out of the base system.
        ++ lib.optionals pkgs.stdenv.hostPlatform.isLinux [
          pkgs.xclip
          pkgs.wl-clipboard
        ];

      # ======================================================================
      # PER-REPO BLOCK 2 -- libraries that get dlopened, not linked
      # ======================================================================
      # Empty, and that is the honest answer for this repo: every dependency in
      # go.mod is pure Go (go-git rather than the git binary, atotto/clipboard
      # rather than libX11, filippo.io/age rather than libsodium), so nothing
      # here dlopens a .so and there is no LD_LIBRARY_PATH to set. The generic
      # machinery below leaves the ambient value untouched when this list is
      # empty -- do not add stdenv.cc.cc.lib "just in case", it only widens the
      # closure and shadows host libraries for anything launched from the shell.
      nativeLibs = pkgs: [ ];

      # ======================================================================
      # PER-REPO BLOCK 3 -- constant environment variables
      # ======================================================================
      # Applied to BOTH surfaces -- the dev shell and every `nix run` wrapper --
      # so a command cannot behave differently depending on how it was invoked.
      envVars = pkgs: {
        # The single most valuable Go setting here. Without it, the `go 1.26.2`
        # directive in go.mod makes Go fetch its own toolchain over the network
        # mid-build; with it you get an instant legible "go.mod requires go >=
        # ... (running go 1.26.5; GOTOOLCHAIN=local)". If go.mod ever outruns
        # nixpkgs, bump flake.lock -- do not unset this.
        GOTOOLCHAIN = "local";
        # Avoids "error obtaining VCS status" in worktrees and agent checkouts
        # owned by another uid. Flags GOFLAGS does not understand are ignored per
        # subcommand, so vet/test/mod tidy still work. Deliberately NOT setting
        # -mod=vendor or -mod=mod: this repo has no vendor/ directory and either
        # value breaks the other kind of repo.
        GOFLAGS = "-buildvcs=false";
        # Load-bearing, and NOT merely the release-build convention out of
        # CLAUDE.md. Nothing in this module imports "C", but Go defaults
        # CGO_ENABLED=1 and then the stdlib pulls the cgo `net`/`os/user`
        # resolvers in via go-git, so it wants a C compiler. mkShell puts
        # stdenv's gcc on PATH while the `nix run` wrappers do not, which
        # reproduced as `nix develop -c go test ./...` passing and
        # `nix run .#test` failing with `cgo: C compiler "gcc" not found` --
        # exactly the two-surface divergence this template exists to prevent.
        # Disabling cgo is the fix that keeps the closure lean; adding gcc would
        # be the fix that keeps cgo. If you need `-race` (which requires cgo),
        # run it inside `nix develop` where the compiler is present:
        # `CGO_ENABLED=1 go test -race ./...`.
        CGO_ENABLED = "0";
        # Bubble Tea v2 downsamples colour from the detected terminal profile.
        # Under `nix run` / `nix develop -c` there is no tty and the golden-file
        # tests force NoColor themselves, so nothing here overrides TERM or
        # COLORTERM -- an inherited value from the caller's real terminal is what
        # `dev-run` wants.
      };

      # ======================================================================
      # PER-REPO BLOCK 4 -- the command map
      # ======================================================================
      # THE single source of truth. It generates `apps` (so `nix run .#test`
      # works), the `dev-*` wrappers on PATH inside the shell, and `dev-help`.
      #
      # No `setup` verb: Go fills its module cache on the first build, so there
      # is nothing to bootstrap. That first build DOES need network (go.sum has
      # 40-odd modules and nothing is vendored) -- hence the "(network)" note on
      # `build` and `test`, which drops away once ~/go/pkg/mod is warm.
      #
      # `text` is bash under `set -euo pipefail`, shellcheck'd at BUILD time, and
      # it runs in the caller's current directory so an agent can test
      # uncommitted edits.
      commands = pkgs: {
        build = {
          # Produces the real artifact at the repo root, where .gitignore
          # already expects it (`/sops-tui`). -trimpath is the reproducible-build
          # convention recorded in CLAUDE.md; CGO_ENABLED=0 comes from envVars so
          # it applies to every verb, not just this one.
          #
          # go build takes its flags BEFORE the package argument, so "$@" lands
          # after the package and is only useful for naming extra packages --
          # pass build flags via GOFLAGS instead.
          description = "(network on first run) compile ./cmd/sops-tui to $REPO_ROOT/sops-tui";
          text = ''
            go build -trimpath -o "$REPO_ROOT/sops-tui" "$REPO_ROOT/cmd/sops-tui" "$@"
          '';
        };
        test = {
          # Package list first, then "$@": `go test` explicitly documents
          # "[packages] [build/test flags & test binary flags]", so
          # `nix run .#test -- -run TestFoo -v` works.
          description = "(network on first run) go test ./... -- sops and age-keygen are on PATH, so the e2e tests do not self-skip";
          text = ''go test ./... "$@"'';
        };
        lint = {
          description = "golangci-lint over the module";
          text = ''golangci-lint run "$@"'';
        };
        fmt = {
          description = "gofmt the tree (rewrites files)";
          text = ''go fmt ./... "$@"'';
        };
        run = {
          # Bubble Tea needs a real terminal: under `nix run` with no tty this
          # exits with an open/ioctl error rather than hanging, which is the
          # failure mode we want. `sops` and the clipboard helpers are already
          # on PATH from the toolchain, but an age key at
          # ~/.config/sops/age/keys.txt and a .sops.yaml in the working
          # directory are the user's to provide -- the startup validator
          # reports both as soft warnings.
          description = "start the TUI (needs a real terminal; run it from a repo that has a .sops.yaml)";
          text = ''go run "$REPO_ROOT/cmd/sops-tui" "$@"'';
        };
      };

      # ======================================================================
      # GENERIC MACHINERY -- byte-identical across the fleet, do not edit
      # ======================================================================

      # Prepend, never assign: a host LD_LIBRARY_PATH may be carrying something
      # the user needs, and clobbering it breaks binaries they launch from here.
      # Linux only -- on darwin the loader variable is DYLD_*, and exporting a
      # Linux-shaped value there is at best useless.
      ldPreamble =
        pkgs:
        lib.optionalString (pkgs.stdenv.hostPlatform.isLinux && nativeLibs pkgs != [ ]) ''
          export LD_LIBRARY_PATH="${lib.makeLibraryPath (nativeLibs pkgs)}''${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"
        '';

      # Every command gets $REPO_ROOT. `nix run` and `nix develop` both start in
      # whatever directory they were invoked from, so a bare relative path
      # silently acts on the wrong tree as soon as an agent works from a
      # subdirectory. Note we do NOT cd there: commands act on the caller's cwd
      # on purpose.
      rootPreamble = ''
        REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
        export REPO_ROOT
      '';

      # One derivation per command, reused by both `apps` and the dev shell, so
      # the two can never diverge. `dev-` prefixed because a bare `test` binary
      # earlier on PATH would shadow the POSIX shell builtin and quietly break
      # every script in the repo that uses it.
      wrappers =
        pkgs:
        lib.mapAttrs (
          name: cmd:
          pkgs.writeShellApplication {
            name = "dev-${name}";
            runtimeInputs = toolchain pkgs;
            runtimeEnv = envVars pkgs;
            meta.description = cmd.description;
            text = ''
              ${rootPreamble}
              ${ldPreamble pkgs}
              ${cmd.text}
            '';
          }
        ) (commands pkgs);

      helpFor =
        pkgs:
        let
          cmds = commands pkgs;
          names = lib.attrNames cmds;
          width = lib.foldl' (a: n: lib.max a (builtins.stringLength n)) 0 names;
          pad = n: n + lib.concatStrings (lib.genList (_: " ") (width - builtins.stringLength n));
          line = n: c: "  dev-${pad n}  ${c.description}";
        in
        pkgs.writeShellApplication {
          name = "dev-help";
          meta.description = "print this repo's command map (works offline)";
          text = ''
            cat <<'EOF'
            ${lib.concatStringsSep "\n" (lib.mapAttrsToList line cmds)}
            EOF
          '';
        };
    in
    {
      # `nix flake show` -- the discovery entrypoint, and deliberately the whole
      # machine-facing contract: every app carries a meta.description, which
      # `nix flake show` prints inline and `nix flake show --json` exposes at
      # .apps.<system>.<name>.description. Pure evaluation, so an agent gets the
      # entire command map in one cheap call without reading a README.
      #
      # Do NOT invent a top-level output for this (`agentManifest` and friends):
      # Nix answers with `warning: unknown flake output '<name>'` on every single
      # `nix flake check`, forever.
      apps = forAllSystems (
        pkgs:
        lib.mapAttrs (name: cmd: {
          type = "app";
          program = "${(wrappers pkgs).${name}}/bin/dev-${name}";
          meta.description = cmd.description;
        }) (commands pkgs)
      );

      # `nix develop` -- the toolchain, plus a dev-<verb> for every app.
      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = toolchain pkgs ++ lib.attrValues (wrappers pkgs) ++ [ (helpFor pkgs) ];

          env = envVars pkgs;

          # Some C extensions and node-gyp addons compile at -O0, where glibc's
          # _FORTIFY_SOURCE becomes a hard error instead of a warning. Harmless
          # here (nothing in this repo uses CGo) and kept for fleet uniformity.
          hardeningDisable = [ "fortify" ];

          shellHook = ''
            # mkShell inherits SOURCE_DATE_EPOCH=315532800 (1980-01-01) from
            # stdenv, and any archive built in here then dies with "ZIP does not
            # support timestamps before 1980".
            unset SOURCE_DATE_EPOCH

            ${rootPreamble}
            ${ldPreamble pkgs}

            # Nothing networked, nothing stateful and nothing interactive above
            # this line, and nothing below it either. No `go mod download`, no
            # cache warming: that would make a cold `nix develop -c go vet ./...`
            # start downloading before it runs anything, on EVERY invocation --
            # the exact failure an unattended agent cannot diagnose.

            # The banner is interactive-only, and this guard is load-bearing:
            # shellHook output lands on the STDOUT of `nix develop -c <cmd>`, so
            # an unguarded echo corrupts anything parsing it
            # (`nix develop -c cat x.json | jq` fails to parse). $- is the only
            # reliable discriminator -- it lacks `i` under `nix develop -c` and
            # has it at an interactive prompt. Do not switch to `[ -t 1 ]`: that
            # leaks the banner the moment an agent harness allocates a pty. Do
            # not test $PS1 (unset in both) or $IN_NIX_SHELL (set in both).
            case $- in
              *i*) echo "sops-tui dev shell -- 'dev-help' for the command map" >&2 ;;
            esac
          '';
        };
      });

      # `nix flake check` -- honest by construction. It realises the toolchain
      # closure (so a typo'd or currently-broken attr fails here) and builds
      # every wrapper, which runs shellcheck over every command text. NEVER add
      # a check that always passes: an agent reads "all checks passed!" as a
      # signal, and a fake check makes `nix flake check` a liar.
      #
      # `go test` is deliberately NOT a check: the module cache is not vendored,
      # so it would need network inside the build sandbox. `dev-test` is the
      # honest home for it.
      checks = forAllSystems (pkgs: {
        toolchain =
          pkgs.runCommand "toolchain-check"
            {
              nativeBuildInputs = toolchain pkgs ++ lib.attrValues (wrappers pkgs);
            }
            ''
              for verb in ${lib.escapeShellArgs (lib.attrNames (commands pkgs))}; do
                command -v "dev-$verb" > /dev/null || {
                  echo "dev-$verb is not on PATH" >&2
                  exit 1
                }
              done
              touch "$out"
            '';
      });

      # `nix fmt` -- formats the *Nix* in this repo; Go code is `dev-fmt`.
      # nixfmt-tree (the treefmt wrapper) rather than bare nixfmt, because bare
      # nixfmt tries to parse every path handed to it and fails on non-Nix files.
      # This file ships already formatted, so `nix fmt` is a no-op.
      formatter = forAllSystems (pkgs: pkgs.nixfmt-tree);
    };
}
