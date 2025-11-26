# Cache Chart Minute-Level Update Summary

## What Changed

The DNS cache hit rate chart now displays data at **minute-level granularity** instead of hour-level.

## Key Improvements

### Time Granularity
- **Before**: 24 hours, 1 data point per hour
- **After**: 60 minutes, 1 data point per minute

### Time Format
- **Before**: `D MMM HH:00` (e.g., "26 Nov 14:00")
- **After**: `HH:mm` (e.g., "14:30")

### Real-time Monitoring
- **Before**: Updates every hour
- **After**: Updates every minute

## Technical Changes

### Backend (Already Implemented)
- `DNSCacheStats` maintains 60-minute rolling window
- Updates history every minute
- Returns 60 data points via API

### Frontend (This Update)
- Modified `CacheMetrics.tsx` to use `ResponsiveLine` directly
- Custom time formatting for minute-level display
- X-axis shows time in `HH:mm` format
- Y-axis shows percentage (0-100%)

## Files Changed

1. `client/src/components/Dashboard/CacheMetrics.tsx` - Chart component
2. `AdGuardHome_cache_minutes.exe` - Compiled binary
3. `test_cache_minutes.ps1` - Test script
4. `CACHE_CHART_MINUTE_UPDATE.md` - Detailed documentation

## Testing

Run the test script:
```powershell
.\test_cache_minutes.ps1
```

Or manually verify:
1. Start AdGuardHome
2. Open http://localhost:3000
3. View Dashboard
4. Check "DNS Cache Hit Rate" card
5. Hover over chart to see minute-level timestamps

## Status

✅ Backend: Complete  
✅ Frontend: Complete  
✅ Compilation: Success  
⏳ Testing: Pending user verification
