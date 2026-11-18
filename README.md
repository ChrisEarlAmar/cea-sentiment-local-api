# Local Sentiment API

A self-hosted HTTP service that accepts an array of sentences and returns only an aggregate sentiment result. It runs as a normal Windows or Linux process: no Docker, virtual machine, cloud API, or GPU is needed after the initial download.

The included setup fetches a quantized ONNX form of [`nlptown/bert-base-multilingual-uncased-sentiment`](https://huggingface.co/nlptown/bert-base-multilingual-uncased-sentiment). The underlying classifier predicts a one-to-five-star product-review rating in English, Dutch, German, French, Spanish, and Italian. This service maps one and two stars to negative, three to neutral, and four and five to positive.

## One-file Windows EXE

The bundled Windows x64 build is `cea-sentiment-local-api.exe`, with a small editable `cea-sentiment-local-api.json` beside it. Double-click the EXE and it starts the API on the port from that JSON file, then opens a local browser tester page. Paste a JSON array of strings into that page and click **Analyze**. Your own software can call the same `POST /api/analyze` endpoint.

The supplied configuration uses port 18080 to avoid common local development ports such as 8080:

```json
{
  "port": 18080
}
```

For example, setting `"port": 19090` makes the tester page and API available at `http://127.0.0.1:19090`. If the file is missing, the EXE recreates the same starter configuration beside itself. It binds only to `127.0.0.1`, so it remains local to the computer.

The EXE contains the model, vocabulary, ONNX Runtime, and tester page. On its first start, it automatically writes those internal resources to `%LOCALAPPDATA%\local-sentiment-api\bundled-v1`; no `models` or `runtime` folder needs to travel with the EXE. You can delete that cache while the service is stopped and the next start recreates it.

It targets 64-bit Windows. ONNX Runtime itself relies on the supported Microsoft Visual C++ x64 Redistributable, which is normally already present on current Windows machines. If Windows reports a missing `MSVCP140.dll` or `VCRUNTIME140.dll`, install the current [official x64 redistributable](https://learn.microsoft.com/en-us/cpp/windows/latest-supported-vc-redist) once and launch the same EXE again.

To stop it, close its console window or press `Ctrl+C`. Add `-no-browser` when a background process should not open the tester page.

## API

`POST /api/analyze`

```json
{
  "sentences": [
    "The staff were excellent and helpful.",
    "I waited far too long for an answer.",
    "The product was okay."
  ]
}
```

```json
{
  "total_analyzed": 3,
  "sentiment": {
    "positive": 33.3,
    "negative": 33.3,
    "neutral": 33.3
  },
  "overall_score": 50,
  "overall_sentiment": "mixed"
}
```

Percentages represent the share of input sentences classified in each group. `overall_score` runs from 0 to 100: 0 means every sentence was negative, 50 means positive and negative counts balance, and 100 means every sentence was positive. `overall_sentiment` is positive or negative when one side leads by at least 20 percentage points; otherwise it is `mixed`.

The API accepts one to 5,000 non-blank sentences per request, up to 5 MiB of JSON. It never returns the input text or per-sentence classifications. Requests run sequentially on the local CPU, so split large imports into batches if you need steadier response times.

`GET /health` returns `{"status":"ok"}`.

## First-time setup

You need internet access only for this setup and for downloading Go modules during the build. Runtime inference reads local files only.

The setup downloads about 170 MB for the quantized model and roughly 150 MB for the cross-platform ONNX Runtime package; only the platform-specific runtime libraries are retained. Plan on at least 1.5 GB available RAM for a comfortable CPU-only process. Performance depends on the CPU and sentence length.

### Build the one-file EXE yourself

On Windows x64, after installing Go and the Visual C++ redistributable:

```powershell
.\scripts\fetch-assets.ps1
.\scripts\prepare-bundle.ps1
$env:CGO_ENABLED = "1"
go test -tags bundled ./cmd/sentiment-server ./internal/...
go build -tags bundled -ldflags "-s -w -extldflags=-static" -o .\cea-sentiment-local-api.exe .\cmd\sentiment-server
```

`prepare-bundle.ps1` copies the downloaded resources into the ignored build-only payload directory. The resulting EXE includes them; do not distribute that payload directory.

Or, on Windows, run `scripts\build-dist.bat`. It creates an ignored `dist\` folder containing the distributable EXE and its editable JSON configuration. If no GCC compiler is already installed, the script downloads a checked, portable x64 GCC toolchain once into the ignored `build-tools\` folder; later builds reuse it. This build-only download is not needed to run or distribute the EXE.

### Windows x64

Install a current 64-bit Go toolchain and the Microsoft Visual C++ 2019 or newer redistributable. `build-dist.bat` uses an installed 64-bit GCC when available, otherwise it downloads a checked portable compiler once into ignored build tooling. You can still use MSYS2 UCRT64 with `mingw-w64-ucrt-x86_64-gcc` if you prefer to manage the compiler yourself.

From PowerShell in the project directory:

```powershell
.\scripts\fetch-assets.ps1
$env:CGO_ENABLED = "1"
go test ./cmd/sentiment-server ./internal/...
go build -o .\bin\sentiment-server.exe .\cmd\sentiment-server
.\bin\sentiment-server.exe
```

Test it from another PowerShell window:

```powershell
Invoke-RestMethod http://127.0.0.1:18080/api/analyze -Method Post -ContentType "application/json" -Body '{"sentences":["I love it.","The service was poor.","It was okay."]}'
```

### Linux x64

Install Go plus a C compiler and common build tools, for example `build-essential` on Debian or Ubuntu. The asset script needs `curl`, `unzip`, and `sha256sum`.

```sh
sh scripts/fetch-assets.sh
CGO_ENABLED=1 go test ./cmd/sentiment-server ./internal/...
CGO_ENABLED=1 go build -o bin/sentiment-server ./cmd/sentiment-server
./bin/sentiment-server
```

Test it:

```sh
curl -X POST http://127.0.0.1:18080/api/analyze \
  -H 'Content-Type: application/json' \
  -d '{"sentences":["I love it.","The service was poor.","It was okay."]}'
```

## Separate-file deployment

The ordinary build can instead run beside these folders. This option is useful during development; use the bundled EXE when you want a one-file distribution.

```text
sentiment-service/
├── sentiment-server.exe          # or sentiment-server on Linux
├── models/
│   ├── sentiment.onnx
│   └── vocab.txt
└── runtime/
    ├── onnxruntime.dll            # Windows
    └── onnxruntime_providers_shared.dll
```

Linux uses `libonnxruntime.so` and `libonnxruntime_providers_shared.so` instead. The process defaults to `127.0.0.1:18080`, so only software on the same machine can call it. To listen on a network interface, pass `-addr 0.0.0.0:18080`; put authentication and TLS in front of it before exposing that address outside a trusted network.

Available flags:

```text
-addr 127.0.0.1:18080
-model models/sentiment.onnx
-vocab models/vocab.txt
-runtime runtime/onnxruntime.dll  # runtime/libonnxruntime.so on Linux
-max-tokens 128
```

## Honest runtime notes

The bundled build is a single EXE for distribution, but it must unpack its internal model and native runtime into the current user's cache before inference. That is how a Windows executable can load the ONNX Runtime DLL without making the user manage sidecar files. The separate-file development build needs the local ONNX model, its vocabulary, and the matching ONNX Runtime dynamic library in its deployment folder.

The model has no internet dependency while it runs, and this service does not send text anywhere. ONNX Runtime's CPU library is local too. It is still a machine-learning model, not a general language-understanding system: sarcasm, domain-specific language, code-switching, and languages outside the model's six training languages can produce poor classifications. Test it against representative feedback before using its score for a consequential decision.

## Tests

`go test ./cmd/sentiment-server ./internal/...` covers request validation, the aggregate-only response, score calculations, WordPiece tokenization, and the model's five-star-to-three-sentiment mapping. Add `-tags bundled` after preparing the embedded assets to test the one-file build variant. The model itself is deliberately not included in the repository.

## Third-party components

- The source in this repository is MIT-licensed.
- The selected NLP Town model is [MIT-licensed](https://huggingface.co/nlptown/bert-base-multilingual-uncased-sentiment).
- ONNX Runtime is [MIT-licensed](https://github.com/microsoft/onnxruntime/blob/main/LICENSE).
- The Go ONNX Runtime binding is [MIT-licensed](https://github.com/yalue/onnxruntime_go).
