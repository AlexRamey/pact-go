$pactDir = "$env:APPVEYOR_BUILD_FOLDER\pact"
# $pactDir = "$env:TEMP\pact"
$exitCode = 0

if ($env:PACT_INTEGRATED_TESTS) {
  Remove-Item env:\PACT_INTEGRATED_TESTS
}

# Set environment
if (!($env:GOPATH)) {
  $env:GOPATH = "c:\go"
}
$env:PACT_BROKER_HOST = "https://test.pact.dius.com.au"
$env:PACT_BROKER_USERNAME = "dXfltyFMgNOFZAxr8io9wJ37iUpY42M"
$env:PACT_BROKER_PASSWORD = "O5AIZWxelWbLvqMd8PkAVycBJh2Psyg1"

if (Test-Path "$pactDir") {
  Write-Verbose "-> Deleting old pact directory"
  rmdir -Recurse -Force $pactDir
}

Write-Verbose "--> Creating ${pactDir}"
New-Item -Force -ItemType Directory $pactDir

Write-Verbose "--> Creating pact-go binary"
go build -o "$pactDir\pact-go.exe" "github.com\pact-foundation\pact-go"

Write-Verbose "--> Creating pact daemon (downloading Ruby binaries)"
$downloadDir = $env:TEMP
$url = "https://github.com/pact-foundation/pact-ruby-standalone/releases/download/v1.51.1/pact-1.51.1-win32.zip"

Write-Verbose "    Downloading $url"
$zip = "$downloadDir\pact.zip"
if (!(Test-Path "$zip")) {
  $downloader = new-object System.Net.WebClient
  $downloader.DownloadFile($url, $zip)
}

Write-Verbose "    Extracting $zip"
Add-Type -AssemblyName System.IO.Compression.FileSystem
[System.IO.Compression.ZipFile]::ExtractToDirectory("$zip", $pactDir)

Write-Verbose "    Moving binaries into position"
Get-ChildItem $pactDir
Get-ChildItem $pactDir/pact

Write-Verbose "--> Running tests"
$packages = go list github.com/pact-foundation/pact-go/... |  where {$_ -inotmatch 'vendor'} | where {$_ -inotmatch 'examples'}
$curDir=$pwd

foreach ($package in $packages) {
  Write-Verbose "    Running tests for $package"
  go test -v $package
  if ($LastExitCode -ne 0) {
    Write-Verbose "    ERROR: Test failed, logging failure"
    $exitCode=1
  }
}

Write-Verbose "--> Testing E2E examples"
Write-Verbose "    Starting pact daemon in background"
Start-Process -FilePath "$pactDir\pact-go.exe" -ArgumentList "daemon -v -l DEBUG"  -RedirectStandardOutput "pacte-2e.log" -RedirectStandardError "pact-e2e-error.log"
$env:PACT_INTEGRATED_TESTS=1

$examples=@("github.com/pact-foundation/pact-go/examples/consumer/goconsumer", "github.com/pact-foundation/pact-go/examples/go-kit/provider", "github.com/pact-foundation/pact-go/examples/mux/provider", "github.com/pact-foundation/pact-go/examples/gin/provider")
foreach ($example in $examples) {
  Write-Verbose "    Installing dependencies for example: $example"
  cd "$env:GOPATH\src\$example"
  go get ./...
  Write-Verbose "    Running tests for $example"
  go test -v .
  if ($LastExitCode -ne 0) {
    Write-Verbose "    ERROR: Test failed, logging failure"
    $exitCode=1
  }
}
cd $curDir

Write-Verbose "    Shutting down pact processes :)"
Stop-Process -Name ruby
Stop-Process -Name pact-go

Write-Verbose "    Done!"
if ($exitCode -ne 0) {
  Write-Verbose "--> Build failed, exiting"
  Exit $exitCode
}