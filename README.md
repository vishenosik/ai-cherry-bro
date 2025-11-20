# ai-cherry-bro
My honest attempt to develop an AI agent that autonomously manages a web browser to perform complex multi-step tasks.

# dependencies
- [Playwright](https://playwright.dev/)

```
go get github.com/playwright-community/playwright-go@v0.5200.1
```

```
go run github.com/playwright-community/playwright-go/cmd/playwright@v0.xxxx.x install --with-deps
# Or
go install github.com/playwright-community/playwright-go/cmd/playwright@v0.xxxx.x
playwright install --with-deps
```

# run

Setup .env config with your credentials

```toml
OPENAI_API_KEY=your_openai_api_key
```

Run 

```bash
go run cmd/browser-agent/main.go
```

# requests

There is a gRPC simple API you can explore in `protos/v1/browser_task.proto`

I recommend using postman to do that.