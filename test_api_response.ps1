# Test API response to see if filter_name is included
# First, let's make a DNS query to trigger a rule match
nslookup baidu.com 127.0.0.1

# Wait a moment for the query to be logged
Start-Sleep -Seconds 2

# Now check the API response (you'll need to authenticate)
Write-Host "Please check the browser developer tools Network tab for /control/querylog response"
Write-Host "Look for the 'rules' array in the response and check if 'filter_name' field is present"
