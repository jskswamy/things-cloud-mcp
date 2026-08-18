package main

import "strings"

// LandingPagePreviewHTML keeps the original landing page intact and applies only
// the disclosure and connection-guide changes under review.
var LandingPagePreviewHTML = buildLandingPagePreviewHTML()
var DocsPagePreviewHTML = buildDocsPagePreviewHTML()

func buildLandingPagePreviewHTML() string {
	page := strings.ReplaceAll(LandingPageHTML, "OAuth 2.0", "OAuth 2.1")

	page = strings.Replace(page, "</head>", `<style>
.client-dropdown-item[data-client="claude-code"],
.client-dropdown-item[data-client="cursor"],
.client-dropdown-item[data-client="windsurf"]{display:none}

.community-note-section{padding:0 0 4px}
.warning-box{background:#FFF8EB;border-radius:var(--radius-sm);padding:20px 24px}
.warning-box .warning-title{font-size:15px;font-weight:700;color:var(--text);margin-bottom:10px}
.warning-box p{font-size:14px;color:var(--text-secondary);line-height:1.65}
.client-dropdown-item img{width:16px;height:16px;object-fit:contain;flex-shrink:0}
.client-selector-trigger img{width:32px;height:32px;object-fit:contain}
#clientLabel img{flex-shrink:0;position:relative;top:1px}
.any-client-intro{margin-bottom:18px;font-size:13px;color:var(--text-secondary);line-height:1.55}
.protocol-values{display:grid;grid-template-columns:1.45fr 1fr 1fr 1fr;gap:0;border:1px solid var(--divider);border-radius:var(--radius-sm);overflow:hidden}
.protocol-value{padding:13px 14px;background:var(--code-bg);min-width:0}
.protocol-value:not(:last-child){border-right:1px solid var(--divider)}
.protocol-value span{display:block;font-size:11px;font-weight:600;color:var(--text-secondary);margin-bottom:4px}
.protocol-value code{display:block;font-family:"SF Mono",SFMono-Regular,Menlo,Monaco,Consolas,monospace;font-size:12px;color:var(--text);overflow-wrap:anywhere}
.protocol-note{margin-top:12px;font-size:12px;color:var(--text-secondary);line-height:1.55}
.faq-section{padding-top:0}
.faq-section .section-header{margin-bottom:22px}
.faq-list{border-top:1px solid var(--divider)}
.faq-item{border-bottom:1px solid var(--divider)}
.faq-item summary{display:flex;align-items:center;justify-content:space-between;gap:20px;padding:18px 0;cursor:pointer;list-style:none;font-size:14px;font-weight:600;color:var(--text)}
.faq-item summary::-webkit-details-marker{display:none}
.faq-item summary::after{content:"+";flex:none;color:var(--text-secondary);font-size:20px;font-weight:400;line-height:1}
.faq-item[open] summary::after{content:"\2212"}
.faq-answer{max-width:760px;padding:0 34px 18px 0;color:var(--text-secondary);font-size:13px;line-height:1.65}
.faq-answer a{color:var(--blue);font-weight:500;text-decoration:none}
.faq-answer a:hover{text-decoration:underline}
@media(prefers-color-scheme:dark){
  .warning-box{background:#3A3520}
}
@media(max-width:700px){
  .protocol-values{grid-template-columns:1fr 1fr}
  .protocol-value:nth-child(2){border-right:0}
  .protocol-value:nth-child(-n+2){border-bottom:1px solid var(--divider)}
}
@media(max-width:480px){
  .warning-box{padding:18px 20px}
  .protocol-values{grid-template-columns:1fr}
  .protocol-value:not(:last-child){border-right:0;border-bottom:1px solid var(--divider)}
}
</style>
</head>`, 1)

	page = strings.Replace(page, "<!-- Connection -->", `<!-- Disclosure -->
<section class="community-note-section" aria-label="About this project">
  <div class="container">
    <div class="warning-box">
      <div class="warning-title">&#9888;&#65039; Unofficial API</div>
      <p>This community project talks directly to the Things Cloud backend using a reverse-engineered API. It is not an official Cultured Code integration. The integration is actively maintained and used by me, but it may stop working if Things changes its backend.</p>
    </div>
  </div>
</section>

<!-- Connection -->`, 1)

	const mcpIcon = `<img src="https://avatars.githubusercontent.com/u/182288589?v=4" alt="" width="16" height="16">`
	page = strings.Replace(page, `<div class="client-dropdown-item" data-client="cursor">`, `<div class="client-dropdown-item" data-client="any-mcp">`+mcpIcon+` Any MCP Client</div><div class="client-dropdown-item" data-client="cursor">`, 1)
	page = strings.Replace(page, `    "cursor":`, `    "any-mcp": '`+mcpIcon+`',
    "cursor":`, 1)
	page = strings.Replace(page, `    "cursor": "Cursor",`, `    "any-mcp": "Any MCP Client",
    "cursor": "Cursor",`, 1)
	page = strings.Replace(page, `<div class="client-dropdown-item active" data-client="claude-ai">`, `<div class="client-dropdown-item" data-client="claude-ai">`, 1)
	page = strings.Replace(page, `<div class="client-dropdown-item" data-client="chatgpt">`, `<div class="client-dropdown-item active" data-client="chatgpt">`, 1)

	oldClaude := `      <div class="client-instructions active" data-client="claude-ai">
        <div class="step"><span class="num">1</span><div class="step-text">Go to <strong>Settings &rarr; Connectors &rarr; Add custom connector</strong></div></div>
        <div class="step"><span class="num">2</span><div class="step-text">Enter name: <strong>Things Cloud</strong></div></div>
        <div class="step"><span class="num">3</span><div class="step-text">Enter URL: <strong><span class="mcp-url"></span></strong></div></div>
        <div class="step"><span class="num">4</span><div class="step-text">Click <strong>Add</strong>, then enable in chat via the &ldquo;+&rdquo; button</div></div>
      </div>`
	newClaude := `      <div class="client-instructions" data-client="claude-ai">
        <div class="step"><span class="num">1</span><div class="step-text">Open <strong>Customize &rarr; Connectors</strong> in Claude.</div></div>
        <div class="step"><span class="num">2</span><div class="step-text">Click <strong>+</strong>, then choose <strong>Add custom connector</strong>.</div></div>
        <div class="step"><span class="num">3</span><div class="step-text">Enter name: <strong>Things Cloud</strong> and URL: <strong><span class="mcp-url"></span></strong>.</div></div>
        <div class="step"><span class="num">4</span><div class="step-text">Click <strong>Add</strong>, then <strong>Connect</strong> and complete the Things Cloud sign-in. Enable it in a chat from <strong>+ &rarr; Connectors</strong>.</div></div>
        <div class="note"><a href="https://support.claude.com/en/articles/11175166-get-started-with-custom-connectors-using-remote-mcp" target="_blank" rel="noopener">View Claude&rsquo;s official instructions &nearr;</a></div>
      </div>`
	page = strings.Replace(page, oldClaude, newClaude, 1)

	oldChatGPT := `      <div class="client-instructions" data-client="chatgpt">
        <div class="step"><span class="num">1</span><div class="step-text">Go to <strong>Settings &rarr; Apps &amp; Connectors &rarr; Advanced</strong>, enable <strong>Developer Mode</strong></div></div>
        <div class="step"><span class="num">2</span><div class="step-text">Click <strong>Add Connector</strong></div></div>
        <div class="step"><span class="num">3</span><div class="step-text">Enter name: <strong>Things Cloud</strong>, URL: <strong><span class="mcp-url"></span></strong></div></div>
        <div class="step"><span class="num">4</span><div class="step-text">In a new chat, click &ldquo;+&rdquo; to select the connector</div></div>
        <div class="note">Note: ChatGPT requires a publicly accessible URL (use ngrok for local dev).</div>
      </div>`
	newChatGPT := `      <div class="client-instructions active" data-client="chatgpt">
        <div class="step"><span class="num">1</span><div class="step-text">Open <strong>Settings &rarr; Security and login</strong> and enable <strong>Developer mode</strong>.</div></div>
        <div class="step"><span class="num">2</span><div class="step-text">Open <strong>Plugins</strong> from the sidebar and click <strong>Create app</strong>.</div></div>
        <div class="step"><span class="num">3</span><div class="step-text">Use name: <strong>Things Cloud</strong>, connection: <strong>Server URL</strong>, URL: <strong><span class="mcp-url"></span></strong>, and authentication: <strong>OAuth</strong>.</div></div>
        <div class="step"><span class="num">4</span><div class="step-text">Acknowledge the unreviewed-server warning, click <strong>Create</strong>, and complete the Things Cloud sign-in.</div></div>
      </div>`
	page = strings.Replace(page, oldChatGPT, newChatGPT, 1)

	page = strings.Replace(page, claudeCodeInstructions, "", 1)
	page = strings.Replace(page, cursorInstructions, "", 1)
	page = strings.Replace(page, windsurfInstructions, "", 1)

	page = strings.Replace(page, `    </div>
  </div>
</section>

<!-- Footer -->`, `
      <div class="client-instructions" data-client="any-mcp">
        <p class="any-client-intro">Use these standard connection details in any MCP client that supports remote Streamable HTTP.</p>
      <div class="protocol-values">
        <div class="protocol-value"><span>Server URL</span><code class="mcp-url"></code></div>
        <div class="protocol-value"><span>Transport</span><code>Streamable HTTP</code></div>
        <div class="protocol-value"><span>Authentication</span><code>OAuth 2.1 + PKCE</code></div>
        <div class="protocol-value"><span>Registration</span><code>DCR</code></div>
      </div>
      <p class="protocol-note">Basic authentication remains available for legacy clients, but OAuth is recommended.</p>
    </div>
  </div>
</section>

<!-- FAQ -->
<section class="faq-section">
  <div class="container">
    <div class="section-header">
      <div class="section-title">Frequently asked questions</div>
    </div>
    <div class="faq-list">
      <details class="faq-item">
        <summary>Does the hosted service receive my Things Cloud credentials?</summary>
        <div class="faq-answer">Yes. The hosted service receives your Things Cloud credentials and mediates your task data. OAuth-connected credentials are encrypted at rest. If you do not want a third-party server in that path, you can <a href="https://github.com/wbopan/things-cloud-mcp" target="_blank" rel="noopener">self-host the project</a>.</div>
      </details>
      <details class="faq-item">
        <summary>Is compatibility guaranteed?</summary>
        <div class="faq-answer">No. The integration is based on reverse engineering, and Things can change its backend without notice. I use and maintain the service myself, but this hobby project has no SLA and may occasionally stop working.</div>
      </details>
      <details class="faq-item">
        <summary>What should I do before connecting?</summary>
        <div class="faq-answer">Keep a recent backup of your Things data and review write or destructive actions before approving them. The tools can create, modify, complete, and delete Things data.</div>
      </details>
    </div>
  </div>
</section>

<!-- Footer -->`, 1)

	page = strings.Replace(page, `var mcpUrl = window.location.origin + "/mcp";`, `var isLocal = window.location.hostname === "localhost" || window.location.hostname === "127.0.0.1";
  var mcpUrl = isLocal ? "https://thingscloudmcp.com/mcp" : window.location.origin + "/mcp";`, 1)
	page = strings.Replace(page, `  var items = dropdown.querySelectorAll(".client-dropdown-item");`, `  var chatGPTItem = dropdown.querySelector('[data-client="chatgpt"]');
  if (chatGPTItem) dropdown.insertBefore(chatGPTItem, dropdown.firstElementChild);
  var items = dropdown.querySelectorAll(".client-dropdown-item");`, 1)
	page = strings.Replace(page, `  label.innerHTML = iconMap["claude-ai"] + " " + nameMap["claude-ai"];`, `  label.innerHTML = iconMap["chatgpt"] + " " + nameMap["chatgpt"];`, 1)

	return page
}

func buildDocsPagePreviewHTML() string {
	page := strings.ReplaceAll(DocsPageHTML, "OAuth 2.0", "OAuth 2.1")

	page = strings.Replace(page, "</head>", `<style>
.client-dropdown-item[data-client="claude-code"],
.client-dropdown-item[data-client="cursor"],
.client-dropdown-item[data-client="windsurf"]{display:none}
.client-dropdown-item img{width:16px;height:16px;object-fit:contain;flex-shrink:0}
.client-selector-trigger img{width:32px;height:32px;object-fit:contain}
#docsClientLabel img{flex-shrink:0;position:relative;top:1px}
.any-client-intro{margin-bottom:18px;font-size:13px;color:var(--text-secondary);line-height:1.55}
.protocol-values{display:grid;grid-template-columns:1.45fr 1fr 1fr 1fr;gap:0;border:1px solid var(--divider);border-radius:var(--radius-sm);overflow:hidden}
.protocol-value{padding:13px 14px;background:var(--code-bg);min-width:0}
.protocol-value:not(:last-child){border-right:1px solid var(--divider)}
.protocol-value span{display:block;font-size:11px;font-weight:600;color:var(--text-secondary);margin-bottom:4px}
.protocol-value code{display:block;font-family:"SF Mono",SFMono-Regular,Menlo,Monaco,Consolas,monospace;font-size:12px;color:var(--text);overflow-wrap:anywhere}
.protocol-note{margin-top:12px;font-size:12px;color:var(--text-secondary);line-height:1.55}
@media(max-width:700px){
  .protocol-values{grid-template-columns:1fr 1fr}
  .protocol-value:nth-child(2){border-right:0}
  .protocol-value:nth-child(-n+2){border-bottom:1px solid var(--divider)}
}
@media(max-width:480px){
  .protocol-values{grid-template-columns:1fr}
  .protocol-value:not(:last-child){border-right:0;border-bottom:1px solid var(--divider)}
}
</style>
</head>`, 1)

	const mcpIcon = `<img src="https://avatars.githubusercontent.com/u/182288589?v=4" alt="" width="16" height="16">`
	page = strings.Replace(page, `<div class="client-dropdown-item" data-client="cursor">`, `<div class="client-dropdown-item" data-client="any-mcp">`+mcpIcon+` Any MCP Client</div><div class="client-dropdown-item" data-client="cursor">`, 1)
	page = strings.Replace(page, `    "cursor":`, `    "any-mcp": '`+mcpIcon+`',
    "cursor":`, 1)
	page = strings.Replace(page, `    "cursor": "Cursor",`, `    "any-mcp": "Any MCP Client",
    "cursor": "Cursor",`, 1)
	page = strings.Replace(page, `<div class="client-dropdown-item active" data-client="claude-ai">`, `<div class="client-dropdown-item" data-client="claude-ai">`, 1)
	page = strings.Replace(page, `<div class="client-dropdown-item" data-client="chatgpt">`, `<div class="client-dropdown-item active" data-client="chatgpt">`, 1)
	page = strings.Replace(page, `id="docsClientLabel">Claude.ai`, `id="docsClientLabel">ChatGPT`, 1)

	page = replaceHTMLBlock(page,
		`    <div class="client-instructions active" data-client="claude-ai">`,
		`    <div class="client-instructions" data-client="claude-code">`,
		`    <div class="client-instructions" data-client="claude-ai">
      <div class="step"><span class="num">1</span><div class="step-text">Open <strong>Customize &rarr; Connectors</strong> in Claude.</div></div>
      <div class="step"><span class="num">2</span><div class="step-text">Click <strong>+</strong>, then choose <strong>Add custom connector</strong>.</div></div>
      <div class="step"><span class="num">3</span><div class="step-text">Enter name: <strong>Things Cloud</strong> and URL: <strong><span class="mcp-url"></span></strong>.</div></div>
      <div class="step"><span class="num">4</span><div class="step-text">Click <strong>Add</strong>, then <strong>Connect</strong> and complete the Things Cloud sign-in. Enable it in a chat from <strong>+ &rarr; Connectors</strong>.</div></div>
      <div class="note"><a href="https://support.claude.com/en/articles/11175166-get-started-with-custom-connectors-using-remote-mcp" target="_blank" rel="noopener">View Claude&rsquo;s official instructions &nearr;</a></div>
    </div>

`)
	page = replaceHTMLBlock(page,
		`    <div class="client-instructions" data-client="claude-code">`,
		`    <div class="client-instructions" data-client="chatgpt">`,
		"")
	page = replaceHTMLBlock(page,
		`    <div class="client-instructions" data-client="chatgpt">`,
		`    <div class="client-instructions" data-client="cursor">`,
		`    <div class="client-instructions active" data-client="chatgpt">
      <div class="step"><span class="num">1</span><div class="step-text">Open <strong>Settings &rarr; Security and login</strong> and enable <strong>Developer mode</strong>.</div></div>
      <div class="step"><span class="num">2</span><div class="step-text">Open <strong>Plugins</strong> from the sidebar and click <strong>Create app</strong>.</div></div>
      <div class="step"><span class="num">3</span><div class="step-text">Use name: <strong>Things Cloud</strong>, connection: <strong>Server URL</strong>, URL: <strong><span class="mcp-url"></span></strong>, and authentication: <strong>OAuth</strong>.</div></div>
      <div class="step"><span class="num">4</span><div class="step-text">Acknowledge the unreviewed-server warning, click <strong>Create</strong>, and complete the Things Cloud sign-in.</div></div>
    </div>

`)
	page = replaceHTMLBlock(page,
		`    <div class="client-instructions" data-client="cursor">`,
		`    <div class="client-instructions" data-client="windsurf">`,
		"")

	const instructionsEnd = `

  </div>
</div>

</div><!-- /container -->`
	page = replaceHTMLBlock(page,
		`    <div class="client-instructions" data-client="windsurf">`,
		instructionsEnd,
		"")
	page = strings.Replace(page, instructionsEnd, `

    <div class="client-instructions" data-client="any-mcp">
      <p class="any-client-intro">Use these standard connection details in any MCP client that supports remote Streamable HTTP.</p>
      <div class="protocol-values">
        <div class="protocol-value"><span>Server URL</span><code class="mcp-url"></code></div>
        <div class="protocol-value"><span>Transport</span><code>Streamable HTTP</code></div>
        <div class="protocol-value"><span>Authentication</span><code>OAuth 2.1 + PKCE</code></div>
        <div class="protocol-value"><span>Registration</span><code>DCR</code></div>
      </div>
      <p class="protocol-note">Basic authentication remains available for legacy clients, but OAuth is recommended.</p>
    </div>`+instructionsEnd, 1)

	page = strings.Replace(page, `var mcpUrl = window.location.origin + "/mcp";`, `var isLocal = window.location.hostname === "localhost" || window.location.hostname === "127.0.0.1";
  var mcpUrl = isLocal ? "https://thingscloudmcp.com/mcp" : window.location.origin + "/mcp";`, 1)
	page = strings.Replace(page, `  var items = dropdown.querySelectorAll(".client-dropdown-item");`, `  var chatGPTItem = dropdown.querySelector('[data-client="chatgpt"]');
  if (chatGPTItem) dropdown.insertBefore(chatGPTItem, dropdown.firstElementChild);
  var items = dropdown.querySelectorAll(".client-dropdown-item");`, 1)
	page = strings.Replace(page, `  label.innerHTML = iconMap["claude-ai"] + " " + nameMap["claude-ai"];`, `  label.innerHTML = iconMap["chatgpt"] + " " + nameMap["chatgpt"];`, 1)

	return page
}

func replaceHTMLBlock(page, startMarker, endMarker, replacement string) string {
	start := strings.Index(page, startMarker)
	if start < 0 {
		return page
	}
	endOffset := strings.Index(page[start:], endMarker)
	if endOffset < 0 {
		return page
	}
	return page[:start] + replacement + page[start+endOffset:]
}

const claudeCodeInstructions = `      <div class="client-instructions" data-client="claude-code">
        <div class="step"><span class="num">1</span><div class="step-text">Run the following command:</div></div>
        <pre><code>claude mcp add --transport http \
  things-cloud <span class="mcp-url"></span> \
  --header "Authorization: Basic BASE64_ENCODE(email:password)"</code></pre>
        <div class="note">Replace <strong>BASE64_ENCODE(email:password)</strong> with your base64-encoded Things Cloud credentials (email:password). Generate with: <code>echo -n 'email:password' | base64</code></div>
        <div class="step"><span class="num">2</span><div class="step-text">Verify with the <strong>/mcp</strong> command inside Claude Code.</div></div>
      </div>`

const cursorInstructions = `      <div class="client-instructions" data-client="cursor">
        <div class="step"><span class="num">1</span><div class="step-text">Add to <strong>~/.cursor/mcp.json</strong>:</div></div>
        <pre><code>{
  "mcpServers": {
    "things-cloud": {
      "url": "<span class="mcp-url"></span>",
      "headers": {
        "Authorization": "Basic BASE64_ENCODE(email:password)"
      }
    }
  }
}</code></pre>
        <div class="note">Replace <strong>BASE64_ENCODE(email:password)</strong> with your base64-encoded Things Cloud credentials (email:password). Generate with: <code>echo -n 'email:password' | base64</code></div>
      </div>`

const windsurfInstructions = `      <div class="client-instructions" data-client="windsurf">
        <div class="step"><span class="num">1</span><div class="step-text">Add to <strong>~/.codeium/windsurf/mcp_config.json</strong>:</div></div>
        <pre><code>{
  "mcpServers": {
    "things-cloud": {
      "serverUrl": "<span class="mcp-url"></span>",
      "headers": {
        "Authorization": "Basic BASE64_ENCODE(email:password)"
      }
    }
  }
}</code></pre>
        <div class="note">Replace <strong>BASE64_ENCODE(email:password)</strong> with your base64-encoded Things Cloud credentials (email:password). Generate with: <code>echo -n 'email:password' | base64</code></div>
      </div>`
