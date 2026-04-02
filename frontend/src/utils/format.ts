/**
 * Shared formatting utilities used across multiple views.
 */

/**
 * Convert a byte count to a human-readable string (e.g. "1.23 GB").
 * Supports units up to TB.
 */
export function formatBytes(bytes: number): string {
    if (!bytes) {
        return '0 B'
    }
    const units = ['B', 'KB', 'MB', 'GB', 'TB']
    const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
    return `${(bytes / 1024 ** index).toFixed(index === 0 ? 0 : 2)} ${units[index]}`
}
