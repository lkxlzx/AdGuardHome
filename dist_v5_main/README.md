# AdGuardHome V5 Main Platforms Build

## Version Information
- Version: v5.0.0
- Build Time: 2025-11-28 11:31:25
- Branch: v5

## Features
- Hot domains tracking fix (uses tracked_domains instead of hot_domains)
- Prefetch UI improvements (always show last prefetch time)
- All V4 features and optimizations

## Build Results

| Platform | OS | Architecture | File | Size | Status |
|----------|----|--------------| -----|------|--------|
| Windows x64 | windows | amd64 | AdGuardHome_windows_amd64.exe | 31.01 MB | Success |
| Windows x86 | windows | 386 | AdGuardHome_windows_386.exe | 29.86 MB | Success |
| Linux x64 | linux | amd64 | AdGuardHome_linux_amd64 | 32.27 MB | Success |
| Linux ARM64 | linux | arm64 | AdGuardHome_linux_arm64 | 30.44 MB | Success |
| macOS x64 | darwin | amd64 | AdGuardHome_darwin_amd64 | 32.37 MB | Success |
| macOS ARM64 | darwin | arm64 | AdGuardHome_darwin_arm64 | 30.78 MB | Success |

## Installation

### Windows
1. Download `AdGuardHome_windows_amd64.exe` (64-bit) or `AdGuardHome_windows_386.exe` (32-bit)
2. Rename to `AdGuardHome.exe`
3. Run `AdGuardHome.exe -s install` to install as service
4. Access web interface at http://localhost:3000

### Linux
1. Download `AdGuardHome_linux_amd64` (x64) or `AdGuardHome_linux_arm64` (ARM64)
2. Make executable: `chmod +x AdGuardHome_linux_*`
3. Run: `./AdGuardHome_linux_*`
4. Access web interface at http://localhost:3000

### macOS
1. Download `AdGuardHome_darwin_amd64` (Intel) or `AdGuardHome_darwin_arm64` (Apple Silicon)
2. Make executable: `chmod +x AdGuardHome_darwin_*`
3. Run: `./AdGuardHome_darwin_*`
4. Access web interface at http://localhost:3000

## Changes in This Build

### Hot Domains Fix
- Fixed hot domains count display issue
- Now uses `tracked_domains` metric for stable count
- Count only decreases during cleanup (not during prefetch scheduling)

### Prefetch UI Improvements
- "Last Prefetch" field now always visible
- Shows "--" when no prefetch has occurred yet
- Better user experience and consistency

## Notes
- All binaries are statically compiled (CGO_ENABLED=0)
- Binaries are stripped for smaller size (-ldflags "-s -w")
- Based on V5 branch with all V4 improvements

## Support
For issues or questions, please visit:
https://github.com/lkxlzx/AdGuardHome

---
Built on: 2025-11-28 11:31:25
