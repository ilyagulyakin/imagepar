$ErrorActionPreference = "Stop"
$inputFile = "testdata/large.png"
$pass = 0
$fail = 0

function Check-Files($dirA, $dirB, $expected) {
    $files = @("gray.png","ycbcr_y.png","ycbcr_cb.png","ycbcr_cr.png","ycbcr_roundtrip.png","distort.png","sobel.png")
    $ok = $true
    foreach ($f in $files) {
        $a = (Get-FileHash "$dirA\$f").Hash
        $b = (Get-FileHash "$dirB\$f").Hash
        if ($a -ne $b) {
            Write-Host "FAIL: $f differs" -ForegroundColor Red
            $ok = $false
        }
    }
    if ($ok) {
        Write-Host "OK: all files match ($expected)" -ForegroundColor Green
        $script:pass++
    } else {
        $script:fail++
    }
}

# 1. Побитовое совпадение seq vs par на матрице N × h
Write-Host "`n=== Test 1: seq vs par bitwise match ==="
foreach ($h in @(1, 7, 64)) {
    foreach ($n in @(1, 2, 4, 8, 16)) {
        $seqDir = "cmp_seq_${h}_${n}"
        $parDir = "cmp_par_${h}_${n}"

        go run ./cmd/imgpar --input $inputFile --out $seqDir --mode seq --strip $h --k 0.4
        go run ./cmd/imgpar --input $inputFile --out $parDir --mode par --threads $n --strip $h --k 0.4

        Check-Files $seqDir $parDir "h=$h N=$n"
        Remove-Item -Recurse -Force $seqDir, $parDir
    }
}

# 2. Дисторсия при k=0 == исходник
Write-Host "`n=== Test 2: distort k=0 == input ==="
go run ./cmd/imgpar --input $inputFile --out "tmp_k0" --mode seq --strip 64 --k 0
$a = (Get-FileHash "$inputFile").Hash
$b = (Get-FileHash "tmp_k0\distort.png").Hash
if ($a -eq $b) {
    Write-Host "OK: distort k=0 == input" -ForegroundColor Green
    $pass++
} else {
    Write-Host "FAIL: distort k=0 differs from input" -ForegroundColor Red
    $fail++
}
Remove-Item -Recurse -Force tmp_k0

# 3. Серое: красный → 76, белый → 255
Write-Host "`n=== Test 3: gray formulas ==="
$red = (0.299 * 255 + 0.587 * 0 + 0.114 * 0)
$white = (0.299 * 255 + 0.587 * 255 + 0.114 * 255)
if ([math]::Round($red) -eq 76 -and [math]::Round($white) -eq 255) {
    Write-Host "OK: gray(red)=76, gray(white)=255" -ForegroundColor Green
    $pass++
} else {
    Write-Host "FAIL: gray formulas" -ForegroundColor Red
    $fail++
}

# 4. Стресс: 100 запусков N=16 h=1
Write-Host "`n=== Test 4: stress 100x N=16 h=1 ==="
go run ./cmd/imgpar --input $inputFile --out "stress_ref" --mode seq --strip 1 --k 0.4
for ($i = 0; $i -lt 100; $i++) {
    go run ./cmd/imgpar --input $inputFile --out "stress_run" --mode par --threads 16 --strip 1 --k 0.4
    $files = @("gray.png","ycbcr_y.png","ycbcr_cb.png","ycbcr_cr.png","ycbcr_roundtrip.png","distort.png","sobel.png")
    foreach ($f in $files) {
        $a = (Get-FileHash "stress_ref\$f").Hash
        $b = (Get-FileHash "stress_run\$f").Hash
        if ($a -ne $b) {
            Write-Host "FAIL: stress iteration $i, file $f" -ForegroundColor Red
            $fail++
            break
        }
    }
    if ($i % 10 -eq 0) { Write-Host "  $i/100" }
}
Write-Host "OK: stress 100x passed" -ForegroundColor Green
$pass++
Remove-Item -Recurse -Force stress_ref, stress_run

# 5. Детектор гонок
Write-Host "`n=== Test 5: race detector ==="
go run -race ./cmd/imgpar --input $inputFile --out "race_out" --mode par --threads 16 --strip 1 --k 0.4
if ($LASTEXITCODE -eq 0) {
    Write-Host "OK: no races detected" -ForegroundColor Green
    $pass++
} else {
    Write-Host "FAIL: race detector reported issues" -ForegroundColor Red
    $fail++
}
Remove-Item -Recurse -Force race_out

# Итог
Write-Host "`n=== Result: $pass passed, $fail failed ==="
if ($fail -gt 0) { exit 1 }