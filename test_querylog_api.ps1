# Test query log API response
$username = "lkxlzx"
$password = "lkxlzx"
$base64AuthInfo = [Convert]::ToBase64String([Text.Encoding]::ASCII.GetBytes(("{0}:{1}" -f $username, $password)))

$headers = @{
    Authorization = "Basic $base64AuthInfo"
}

# Get query log
$response = Invoke-RestMethod -Uri "http://localhost:80/control/querylog?limit=1" -Headers $headers -Method Get

# Output the response
$response | ConvertTo-Json -Depth 10
