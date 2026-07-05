# Contract: Generated Docker Run Refresh Semantics

## Applies To

All generated Docker run commands shown by the product UI.

## Required Semantic Steps

1. Best-effort remove any existing container with the generated container name.
2. Best-effort remove the old local image referenced by the command.
3. Pull the target image.
4. Run the container with the expected flags, volumes, and environment values.

## Behavior Rules

- Missing container during cleanup must not stop execution.
- Missing image during cleanup must not stop execution.
- Pull failure must stop execution and surface an error.
- Container creation/start failure must stop execution and surface an error.
- The command must remain copy-pastable as one generated operator command.

## Example Semantic Shape

```bash
docker rm -f <container_name> >/dev/null 2>&1 || true
docker image rm -f <image> >/dev/null 2>&1 || true
docker pull <image>
docker run ...
```

## Compatibility Rules

- The refreshed command must produce the same runtime configuration as the
  existing generated command aside from cleanup/pull behavior.
- All generated Docker run variants must share the same refresh semantics.
- Compose examples are not governed by this contract.
