
1. 系統架構：
    - 主要服務：使用 Go 開發的 API 服務
    - 儲存服務：MinIO 作為物件儲存
    - 文件處理：Unstructured API 用於 PDF 轉換
    - 翻譯服務：Ollama 本地部署
    - 資料庫：Postgresql 儲存資源和任務關係
    - 消息佇列：先用 Watermill + Go Channel，後續改用 RabbitMQ

2. API 端點（[Open API](pdf_api.yaml)）：
    - PDF 上傳
    - 資源列表查詢
    - 文件轉換觸發
    - 文件翻譯觸發
    - 任務狀態查詢
    - 任務取消

3. 資源管理：
    - PDF 原始檔案
    - 轉換後的文字檔案（包含中間的 chunks）
    - 翻譯後的文字檔案
    - MinIO 中對應不同 bucket

4. 資料結構：
    - 資源表：追踪所有文件（PDF、文字、翻譯）
    - Chunks 表：追踪文件片段
    - 任務表：追踪轉換和翻譯任務

5. 工作流程：
   a. PDF 處理流程：
    - 上傳 PDF 到 MinIO
    - 使用 Unstructured API 轉換
    - 接收並儲存轉換後的 chunks
    - 合併 chunks 為完整文字檔

   b. 翻譯流程：
    - 讀取原文 chunks
    - 使用 Ollama 進行翻譯
    - 儲存翻譯後的 chunks
    - 合併為完整翻譯文件
