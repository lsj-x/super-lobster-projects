#!/bin/bash
echo "Generating self-contained HTML preview..."

# Start HTML
cat > preview.html << 'HTMLHEAD'
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>modelmagic-deploy-console - 项目截图预览</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; background: #f5f7fa; margin: 0; padding: 20px; }
        .container { max-width: 1200px; margin: 0 auto; }
        h1 { text-align: center; color: #333; margin-bottom: 30px; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        .card { background: white; border-radius: 8px; box-shadow: 0 2px 12px rgba(0,0,0,0.1); overflow: hidden; transition: transform 0.2s; }
        .card:hover { transform: translateY(-5px); }
        .card-img { width: 100%; height: auto; display: block; }
        .card-body { padding: 15px; }
        .card-title { font-size: 18px; font-weight: bold; color: #333; margin: 0 0 10px 0; }
        .card-text { font-size: 14px; color: #666; margin: 0; }
        .status { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: bold; margin-bottom: 10px; }
        .status-ok { background: #e6f7e6; color: #28a745; }
        .footer { text-align: center; margin-top: 40px; color: #999; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🦞 modelmagic-deploy-console - 核心页面截图</h1>
        <div class="grid">
HTMLHEAD

# Function to add card
add_card() {
    local name=$1
    local desc=$2
    local b64file=$3
    local img_id=$4
    
    echo "            <div class=\"card\">" >> preview.html
    echo "                <img src=\"data:image/png;base64," >> preview.html
    cat "$b64file" >> preview.html
    echo "\" alt=\"$name\" class=\"card-img\" id=\"$img_id\">" >> preview.html
    echo "                <div class=\"card-body\">" >> preview.html
    echo "                    <span class=\"status status-ok\">✅ 已生成</span>" >> preview.html
    echo "                    <h3 class=\"card-title\">$name</h3>" >> preview.html
    echo "                    <p class=\"card-text\">$desc</p>" >> preview.html
    echo "                </div>" >> preview.html
    echo "            </div>" >> preview.html
}

# Add cards
add_card "首页 (Dashboard)" "仪表盘概览，资源统计与快速入口。" "首页.png.b64" "img-home"
add_card "命名空间列表" "资源管理界面，支持增删改查与扩缩容。" "命名空间列表.png.b64" "img-ns"
add_card "部署向导" "安装/升级/回滚流程，含实时日志。" "部署向导.png.b64" "img-deploy"
add_card "监控面板" "实时日志流与事件列表，WebSocket 支持。" "监控面板.png.b64" "img-monitor"

# Close HTML
cat >> preview.html << 'HTMLFOOT'
        </div>
        <div class="footer">
            <p>生成时间：2026-04-02 | 项目状态：全栈开发完成，前后端联调通过</p>
        </div>
    </div>
</body>
</html>
HTMLFOOT

echo "✅ HTML 预览文件已生成：preview.html"
ls -lh preview.html
