#!/usr/bin/env python3
with open('/root/.openclaw/workspace/repos/modelmagic-deploy-console/backend/scripts/mmctl.sh', 'r') as f:
    lines = f.readlines()

# 找到并修改包含 " 91 - Scale" 的行
for i, line in enumerate(lines):
    if 'echo " 91 - Scale"' in line and 'Scale (手动' not in line:
        lines[i] = '    echo " 91 - Scale (手动扩缩容)"\n'
        lines.insert(i+1, '    echo " 92 - Enable HPA (自动扩缩容)"\n')
        lines.insert(i+2, '    echo " 93 - Disable HPA"\n')
        lines.insert(i+3, '    echo " 94 - Get HPA Status"\n')
        break

with open('/root/.openclaw/workspace/repos/modelmagic-deploy-console/backend/scripts/mmctl.sh', 'w') as f:
    f.writelines(lines)
print('Done')
