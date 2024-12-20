# Notes of the development process of the project

## 目的
這個專案是 Go 版本的 [Anthropic's Contextual Retriveval](https://www.anthropic.com/news/contextual-retrieval) 內容的實作。

會做這個單純是為了我個人的學習，費曼說過 
> What I cannot create, I do not understand.

，所以我想透過實作這個專案來確定自己對於文章內容掌握程度。

當然還有受到其他文章啟發，例如
- [Building LLM-powered applications in Go](https://go.dev/blog/llmpowered)
- [Building effective agents](https://www.anthropic.com/research/building-effective-agents)

這個專案預計要實作
1. 一個 RAG 檢索器，可以透過 grpc 來上傳檔案和檢索。
2. 一個 MCP 伺服器，讓 LLM 可以模型可以透過這個檢索器取得檔案內容來推理回覆。
3. 一個簡單評估 RAG 檢索器的工具，可以透過這個工具來評估檢索器的效能。

初期階段會先實作基本 RAG 檢索器和評估工具，之後使用 Contextual Retriveval 的技術來優化檢索器。
再來實作 MCP 檢索器，最後根據效能來優化，例如為新增檔案功能新增 Queue 排隊、檢視索引進度和使用框架（genkit or langchainGo）重構專案。

實作過程會使用 Cline、Github Copilot 和 Claude 來加速開發，展示 AI 工具在開發上的實際使用，讓人看到實際 AI 協助開發並不完全像是 Youtube 影片上那樣的完美。

## 規格
專案使用 Go 語言，並且透過 Docker compose 管理依賴服務。
預計使用 
- [Weaviate](https://webstore.ansi.org/standards/iso/isoiec98992024) Go 實作的向量資料庫
- [Ollama]() 負責提供模型用作 embedding
- [Elasticsearch]() 負責提供檔案索引


