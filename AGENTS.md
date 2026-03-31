# AGENTS.md

## WSL Environment

When running this project inside WSL, initialize the shell environment before using repository tools:

```bash
source ~/.zshrc
```

This project relies on environment variables, PATH additions, and toolchain setup that are already configured in `.zshrc`. If you skip this step, many system-installed tools may appear unavailable even though they are already installed on the machine.

## Recommendation

Before running build, dev, or maintenance commands in WSL, start the shell session with:

```bash
source ~/.zshrc && cd /home/cjxh/ClawGame
```

## Documentation First

When changing development plans or adjusting product functionality, consider updating documentation first.

- Product documentation and backend documentation live under the `docs/` directory.
- English documentation is the primary source of truth.
- For new documentation or documentation changes, update the English version first.
- After the English documentation is ready, sync the same change to the corresponding Chinese documentation.

If a feature change affects product behavior, API contracts, backend design, or developer workflow, document it in `docs/` before or alongside the implementation work.
