export const agentImage = "ghcr.io/wdmjieyao/beszel-agent:edge"

export interface RefreshDockerRunSpec {
	containerName: string
	image: string
	runArgs: string[]
}

export function buildRefreshDockerRunSpec(spec: RefreshDockerRunSpec) {
	return spec
}

export function buildRefreshDockerRunCommand(spec: RefreshDockerRunSpec) {
	return [
		`docker rm -f ${spec.containerName} >/dev/null 2>&1 || true`,
		`docker image rm -f ${spec.image} >/dev/null 2>&1 || true`,
		`docker pull ${spec.image}`,
		`docker run ${spec.runArgs[0] ?? ""} --name ${spec.containerName} ${spec.runArgs.slice(1).join(" ")} ${spec.image}`
			.replace(/\s+/g, " ")
			.trim(),
	].join(" && ")
}
