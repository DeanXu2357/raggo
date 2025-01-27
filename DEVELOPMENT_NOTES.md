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

## 規格
專案使用 Go 語言，並且透過 Docker compose 管理依賴服務。
預計使用 
- [Weaviate](https://webstore.ansi.org/standards/iso/isoiec98992024) Go 實作的向量資料庫
- [Ollama]() 負責提供模型用作 embedding
- [Elasticsearch]() 負責提供檔案索引


## Log
### Git commit `177dabe83ecc5bf617a302964c3b52aa56ce0958`
完成簡單 RAG 實現，在 evaluate command 中可以載入檔案和檢索檔案。  
這個版本只是簡單把載入的資料列透過 ollama 用 `nomic-embed-text` embedding 後存到 Weaviate 中，然後透過 Weaviate 的檢索功能來檢索。 

```
$ go run main.go evaluate -i ./data/codebase_chunks.json -e ./data/evaluation_set.jsonl -k 5
Created schema: CodeChunk_1734936252
Successfully imported 737 code chunks to CodeChunk_1734936252
Evaluation Results (k=5):
Total evaluations: 248
Average score: 68.69%
Successfully deleted class CodeChunk_1734936252
$ go run main.go evaluate -i ./data/codebase_chunks.json -e ./data/evaluation_set.jsonl -k 10
Created schema: CodeChunk_1734936262
Successfully imported 737 code chunks to CodeChunk_1734936262
Evaluation Results (k=10):
Total evaluations: 248
Average score: 75.81%
Successfully deleted class CodeChunk_1734936262
$ go run main.go evaluate -i ./data/codebase_chunks.json -e ./data/evaluation_set.jsonl -k 20
Created schema: CodeChunk_1734936271
Successfully imported 737 code chunks to CodeChunk_1734936271
Evaluation Results (k=20):
Total evaluations: 248
Average score: 79.67%
Successfully deleted class CodeChunk_1734936271
```

可以看到分數對比 Anthropic 文章取得的分數有很大的差距。
```
Evaluating retrieval: 100%|██████████| 248/248 [00:06<00:00, 40.70it/s]
Pass@5: 80.92%
Total Score: 0.8091877880184332
Total queries: 248
Evaluating retrieval: 100%|██████████| 248/248 [00:06<00:00, 39.50it/s]
Pass@10: 87.15%
Total Score: 0.8714957757296468
Total queries: 248
Evaluating retrieval: 100%|██████████| 248/248 [00:06<00:00, 39.43it/s]
Pass@20: 90.06%
Total Score: 0.9006336405529954
Total queries: 248
```

不確定影響分數的原因，可能是因為使用的 embedding 模型不同，或是檢索的方式不同。


### Git commit `f45fa034771516b407dc2534bcdbb2d055871460`

加入上下文資料與分塊一同 embedding 。
上下文統整使用 llama3.2:3b 模型，分塊統整使用 nomic-embed-text 模型。

```bash
$ go run main.go evaluate -i ./data/codebase_chunks.json -e ./data/evaluation_set.jsonl -k 5 -c true
Starting evaluation with:
- Input file: ./data/codebase_chunks.json
- Evaluation file: ./data/evaluation_set.jsonl
- k: 5
- Using contextual information: true
- LLM model: llama3.2:3b
Created schema: CodeChunk_1734940640
Importing chunks 100% [========================================] (737/737)
Successfully imported 737 code chunks to CodeChunk_1734940640
Evaluating queries 100% [========================================] (248/248)        
Evaluation Results (k=5):
Total evaluations: 248
Average score: 75.15%
Successfully deleted class CodeChunk_1734940640
$ go run main.go evaluate -i ./data/codebase_chunks.json -e ./data/evaluation_set.jsonl -k 10 -c true
Starting evaluation with:
- Input file: ./data/codebase_chunks.json
- Evaluation file: ./data/evaluation_set.jsonl
- k: 10
- Using contextual information: true
- LLM model: llama3.2:3b
Created schema: CodeChunk_1734941032
Importing chunks 100% [========================================] (737/737)Successfully imported 737 code chunks to CodeChunk_1734941032
Evaluating queries 100% [========================================] (248/248)        
Evaluation Results (k=10):
Total evaluations: 248
Average score: 81.52%
Successfully deleted class CodeChunk_1734941032
$ go run main.go evaluate -i ./data/codebase_chunks.json -e ./data/evaluation_set.jsonl -k 20 -c true
Starting evaluation with:
- Input file: ./data/codebase_chunks.json
- Evaluation file: ./data/evaluation_set.jsonl
- k: 20
- Using contextual information: true
- LLM model: llama3.2:3b
Created schema: CodeChunk_1734941309
Importing chunks 100% [========================================] (737/737)
Successfully imported 737 code chunks to CodeChunk_1734941309
Evaluating queries 100% [========================================] (248/248)        
Evaluation Results (k=20):
Total evaluations: 248
Average score: 86.96%
Successfully deleted class CodeChunk_1734941309
```

可以看到分數有明顯的提升，但是還是沒有 Anthropic 文章的分數高。

### Git commit `1538cba4838d290ea4d2c93c6fd9c76dec21b117`
```bash
$ go run main.go evaluate -i ./data/codebase_chunks.json -e ./data/evaluation_set.jsonl -k 5 -c true --bm25
Starting evaluation with:
- Input file: ./data/codebase_chunks.json
- Evaluation file: ./data/evaluation_set.jsonl
- k: 5
- Using contextual information: true
- LLM model: llama3.2:3b
- Using BM25 scoring: true
Created schema: CodeChunk_1735009800
Importing chunks 100% [========================================] (737/737)              
Successfully imported 737 code chunks to Weaviate
Successfully imported 737 code chunks to Elasticsearch
Evaluating queries 100% [========================================] (248/248)        
Evaluation Results (k=5):
Total evaluations: 248
Average score: 77.42%
Percentage of results from Weaviate: 83.39%
Percentage of results from Elasticsearch: 83.39%
Successfully deleted class CodeChunk_1735009800
$ go run main.go evaluate -i ./data/codebase_chunks.json -e ./data/evaluation_set.jsonl -k 10 -c true --bm25
Starting evaluation with:
- Input file: ./data/codebase_chunks.json
- Evaluation file: ./data/evaluation_set.jsonl
- k: 10
- Using contextual information: true
- LLM model: llama3.2:3b
- Using BM25 scoring: true
Created schema: CodeChunk_1735010137
Importing chunks 100% [========================================] (737/737)              
Successfully imported 737 code chunks to Weaviate
Successfully imported 737 code chunks to Elasticsearch
Evaluating queries 100% [========================================] (248/248)        
Evaluation Results (k=10):
Total evaluations: 248
Average score: 84.71%
Percentage of results from Weaviate: 86.75%
Percentage of results from Elasticsearch: 73.00%
Successfully deleted class CodeChunk_1735010137
$ go run main.go evaluate -i ./data/codebase_chunks.json -e ./data/evaluation_set.jsonl -k 20 -c true --bm25
Starting evaluation with:
- Input file: ./data/codebase_chunks.json
- Evaluation file: ./data/evaluation_set.jsonl
- k: 20
- Using contextual information: true
- LLM model: llama3.2:3b
- Using BM25 scoring: true
Created schema: CodeChunk_1735010384
Importing chunks 100% [========================================] (737/737)              
Successfully imported 737 code chunks to Weaviate
Successfully imported 737 code chunks to Elasticsearch
Evaluating queries 100% [========================================] (248/248)        
Evaluation Results (k=20):
Total evaluations: 248
Average score: 88.78%
Percentage of results from Weaviate: 90.25%
Percentage of results from Elasticsearch: 58.14%
Successfully deleted class CodeChunk_1735010384
```

對比加上前後文  
k=5: 75.15% -> 77.42%
k=10: 81.52% -> 84.71%
k=20: 86.96% -> 88.78%

可以看到普遍有 2-3% 的提升

