import { execSync } from "node:child_process"

/* Rebuild movies_e2e so every run starts from the seed. CI points make at
 * its Postgres service by setting PSQL (and PGPASSWORD) in the env. */
export default function globalSetup() {
	execSync("make e2e-db", { stdio: "inherit" })
}
