/** Append new records onto prev with gap detection. Converts string `created` values to ms timestamps in place.
 * Pass `maxLen` to cap the result length in one copy instead of slicing again after the call. */
export function appendData<T extends { created: string | number | null }>(
	prev: T[],
	newRecords: T[],
	expectedInterval: number,
	maxLen?: number
): T[] {
	if (!newRecords.length) return prev
	// Pre-trim prev so the single slice() below is the only copy we make
	const trimmed = maxLen && prev.length >= maxLen ? prev.slice(-(maxLen - newRecords.length)) : prev
	const result = trimmed.slice()
	let prevTime = (trimmed.at(-1)?.created as number) ?? 0
	for (const record of newRecords) {
		if (record.created !== null) {
			if (typeof record.created === "string") {
				record.created = new Date(record.created).getTime()
			}
			if (prevTime && (record.created as number) - prevTime > expectedInterval * 1.5) {
				result.push({ created: null, ...("stats" in record ? { stats: null } : {}) } as T)
			}
			prevTime = record.created as number
		}
		result.push(record)
	}
	return result
}
