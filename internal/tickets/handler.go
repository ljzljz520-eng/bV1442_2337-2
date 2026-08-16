package tickets

import (
	"encoding/json"
	"io"
	"net/http"
)

type Handler struct {
	cleaner TextCleaner
}

func NewHandler(cleaner TextCleaner) http.Handler {
	h := &Handler{cleaner: cleaner}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.index)
	mux.HandleFunc("POST /api/tickets/clean", h.clean)
	return mux
}

type cleanRequest struct {
	Description string `json:"description"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, indexPage)
}

func (h *Handler) clean(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		writeError(w, http.StatusBadRequest, "请求必须使用 JSON")
		return
	}
	var request cleanRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式无效")
		return
	}
	result, err := cleanDescription(h.cleaner, request.Description)
	if err != nil {
		if err == ErrDescriptionEmpty {
			writeError(w, http.StatusBadRequest, ErrDescriptionEmpty.Error())
			return
		}
		if err == ErrCleanerUnavailable {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "文本清理失败")
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: message})
}

const indexPage = `<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>售后工单描述清理</title>
<style>
:root { color-scheme: light; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #f5f7fa; color: #172033; }
body { margin: 0; min-height: 100vh; }
main { width: min(760px, calc(100% - 32px)); margin: 0 auto; padding: 40px 0 64px; }
h1 { margin: 0 0 8px; font-size: 28px; }
.intro { margin: 0 0 28px; color: #586174; }
form, .result { background: #fff; border: 1px solid #dfe4ec; border-radius: 8px; padding: 24px; box-shadow: 0 4px 18px rgba(23, 32, 51, .06); }
label { display: block; font-weight: 600; margin-bottom: 10px; }
textarea { box-sizing: border-box; width: 100%; min-height: 164px; padding: 12px; border: 1px solid #cbd2dd; border-radius: 6px; resize: vertical; font: inherit; }
textarea:focus { outline: 2px solid #95b8ff; border-color: #4478e8; }
.actions { display: flex; align-items: center; gap: 12px; margin-top: 16px; }
button { border: 0; border-radius: 6px; background: #245fd1; color: #fff; padding: 11px 18px; font: inherit; font-weight: 600; cursor: pointer; }
button:disabled { cursor: wait; opacity: .65; }
.loading { color: #245fd1; }
.error { margin: 18px 0 0; color: #b42318; }
.result { margin-top: 20px; }
.result[hidden], .error[hidden], .loading[hidden] { display: none; }
.cleaned { white-space: pre-wrap; margin: 0 0 20px; color: #26334d; }
.stats { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin: 0; }
.stats div { padding: 14px; background: #f3f6fb; border-radius: 6px; }
dt { color: #586174; font-size: 13px; }
dd { margin: 5px 0 0; font-size: 22px; font-weight: 700; }
@media (max-width: 540px) { main { padding-top: 24px; } form, .result { padding: 18px; } .stats { grid-template-columns: 1fr; } }
</style>
</head>
<body>
<main>
<h1>售后工单描述清理</h1>
<p class="intro">粘贴工单描述，得到清理后的内容与问题分类统计。</p>
<form id="clean-form">
<label for="description">工单描述</label>
<textarea id="description" name="description" required placeholder="请输入客户反馈"></textarea>
<div class="actions">
<button id="submit" type="submit">清理工单</button>
<span id="loading" class="loading" hidden>处理中...</span>
</div>
<p id="error" class="error" role="alert" hidden></p>
</form>
<section id="result" class="result" aria-live="polite" hidden>
<h2>处理结果</h2>
<p id="cleaned" class="cleaned"></p>
<dl class="stats">
<div><dt>字符数</dt><dd id="characters">0</dd></div>
<div><dt>词语数</dt><dd id="words">0</dd></div>
<div><dt>问题分类</dt><dd id="category">其他</dd></div>
</dl>
</section>
</main>
<script>
const form = document.querySelector('#clean-form');
const description = document.querySelector('#description');
const submit = document.querySelector('#submit');
const loading = document.querySelector('#loading');
const error = document.querySelector('#error');
const result = document.querySelector('#result');
form.addEventListener('submit', async (event) => {
  event.preventDefault();
  error.hidden = true;
  result.hidden = true;
  submit.disabled = true;
  loading.hidden = false;
  try {
    const response = await fetch('/api/tickets/clean', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({description: description.value})
    });
    const payload = await response.json();
    if (!response.ok) throw new Error(payload.error || '处理失败');
    document.querySelector('#cleaned').textContent = payload.cleaned_text;
    document.querySelector('#characters').textContent = payload.statistics.characters;
    document.querySelector('#words').textContent = payload.statistics.words;
    document.querySelector('#category').textContent = payload.statistics.category;
    result.hidden = false;
  } catch (requestError) {
    error.textContent = requestError.message;
    error.hidden = false;
  } finally {
    submit.disabled = false;
    loading.hidden = true;
  }
});
</script>
</body>
</html>`
