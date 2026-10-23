# Handoff: deploy user_auth to Cloud Run and Neon

Written 2026-10-22 when the work was paused. Everything below is state a fresh session cannot read from the code.

## Where things stand

- The app, tests, Docker image, CI and docs are done and merged to `main` (PRs #3 to #10; history in `git log`).
  The default branch on GitHub is `main`; `master` is gone from the settings but may still exist as a branch, and the
  empty PR #13 (`master` to `main`) was recommended for closing, not yet closed.
- Open PR #14 (`fix/favicon`) adds an SVG favicon. Not merged. The UI is embedded in the Go binary, so a server
  started before the merge shows Chrome's globe icon.
- Local branch `feature/cloud-run-terraform` holds three unpushed commits: Neon Terraform (`infra/neon/`), Cloud Run
  Terraform (`infra/cloudrun/`) and the `make tf-*` targets with a `make help` default goal. No PR yet.
- Nothing has been applied. No Neon project `user-auth`, no Cloud Run service, no secret exists yet.
- The working tree has the user's own uncommitted move of `DECISIONS.md` to `docs/DECISIONS.md`. Leave it alone and
  do not commit it unless asked; README still links the old path.

## Decisions made this session

- Vercel is not an option: the app is a long-running Go server with Postgres, and Vercel Hobby only runs
  short-lived functions.
- Target: Cloud Run (free allowance, scale to zero) plus Neon free Postgres. The user's GCP project is
  `task-tracker-app-1-a4a61` (Cloud Run already enabled there), region `us-east4`, chosen close to Neon
  `aws-us-east-2`. The project and region are Terraform variables, so they can change.
- `terraform apply` creates cloud resources, so per the user's rules show the commands and wait for their go-ahead.
- `make tf-apply` runs Neon first, then Cloud Run, passing the Neon connection string between the two roots as
  `TF_VAR_database_url`. It is two applies in sequence, not one dependency graph. The first Cloud Run apply uses
  Google's hello image; CI is meant to replace it (`ignore_changes` on the image).
- The database URL lands in `infra/cloudrun`'s local, git-ignored state file. Acceptable for a demo; a remote state
  bucket is the upgrade.
- The user's Chrome shows many `BG.js` errors and a 401 on `/api/v1/me`. Neither is a bug: the first comes from a
  browser extension, the second is the normal signed-out check.

## Next steps, in order

1. User merges PR #14, closes PR #13, and pushes `feature/cloud-run-terraform` (ask before pushing and opening a PR).
2. User runs `make tf-init`, `make tf-plan`, then `make tf-apply`. It needs `NEON_API_KEY` in the environment and
   `gcloud auth application-default login`. `make tf-plan` fails for Cloud Run until Neon has been applied once.
3. Add a CI deploy job to `.github/workflows/ci.yml`: on push to `main`, after the existing jobs pass, build the
   Docker image, push it to the Artifact Registry repo (`terraform -chdir=infra/cloudrun output image_repository`) and
   deploy it to the `user-auth` Cloud Run service. Use keyless GitHub OIDC (Workload Identity Federation); this needs
   one-time `gcloud` setup that the user runs themselves.
4. Add a short README section on deployment and a `DECISIONS.md` entry for the Cloud Run plus Neon choice.

## Suggested skills

- `agent-skills:ci-cd-and-automation` for the deploy job.
- `fullstack-dev-skills:terraform-engineer` for reviewing or extending the Terraform.
- `fullstack-dev-skills:devops-engineer` if the OIDC setup needs more than the basics.
