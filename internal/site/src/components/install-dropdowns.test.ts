import assert from "node:assert/strict"
import { describe, it } from "node:test"
import { agentImage, buildRefreshDockerRunCommand, buildRefreshDockerRunSpec } from "./install-dropdowns-utils.ts"

describe("buildRefreshDockerRunCommand", () => {
	it("removes old container and image, pulls the latest image, then runs with original args", () => {
		const command = buildRefreshDockerRunCommand(
			buildRefreshDockerRunSpec({
				containerName: "beszel-agent",
				image: agentImage,
				runArgs: ["-d", "--network host", "--restart unless-stopped", '-e KEY="pub"', '-e TOKEN="tok"'],
			})
		)

		assert.match(command, /docker rm -f beszel-agent/)
		assert.match(command, /docker image rm -f ghcr\.io\/wdmjieyao\/beszel-agent:edge/)
		assert.match(command, /docker pull ghcr\.io\/wdmjieyao\/beszel-agent:edge/)
		assert.match(
			command,
			/docker run -d --name beszel-agent --network host --restart unless-stopped -e KEY="pub" -e TOKEN="tok" ghcr\.io\/wdmjieyao\/beszel-agent:edge/
		)
	})

	it("reuses the shared builder for all current docker run copy surfaces", () => {
		const addSystemCommand = buildRefreshDockerRunCommand(
			buildRefreshDockerRunSpec({
				containerName: "beszel-agent",
				image: agentImage,
				runArgs: ["-d", "--network host", "--restart unless-stopped"],
			})
		)
		const tokenCommand = buildRefreshDockerRunCommand(
			buildRefreshDockerRunSpec({
				containerName: "beszel-agent",
				image: agentImage,
				runArgs: ["-d", "--network host", "--restart unless-stopped"],
			})
		)

		assert.equal(addSystemCommand, tokenCommand)
	})
})
