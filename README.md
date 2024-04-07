# Dcard
Backend Intern Assignment

設計圖:


壓力測試圖:


設計想法:
Backend Intern Assignment中的Public API由於有10,000個RPS的效能要求，所以降低IO時間跟搜尋時間會是主要目標，變數為查詢參數，選擇單層cache機制，key為條件參數組合，value為符合條件的資料，實作以Redis作為cache的工具，Database只負責儲存所有廣告。

Admin API:
負責新增廣告，有可能因為投放的廣告時間範圍有now造成cache資料不正確，所以如果投放廣告時間範圍有now就會清理cache。

Public API:
負責依照參數條件查詢時間範圍有now的廣告，先以條件參數查詢cache是否有資料，若有則回傳，無則查詢database然後建立cache。


