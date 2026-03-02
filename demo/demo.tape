Output demo/demo.gif

Set Width 1200
Set Height 600
Set FontSize 14
Set Theme "Catppuccin Mocha"
Set Framerate 30
Set PlaybackSpeed 0.8

Hide
Type "go build -o /tmp/git-cx . && echo BUILD_OK"
Enter
Wait+Screen@60s /BUILD_OK/
Type "bash demo/setup.sh"
Enter
Wait+Screen@20s /Setup complete:/
Type "export PATH=/tmp:$PATH"
Enter
Type "cd /tmp/git-cx-demo"
Enter
Sleep 0.5s

# 画面とバッファをクリア（clearコマンドの代わりにCtrl+Lを使用）
Ctrl+L
Sleep 1s
Show

Type "git diff --cached --stat"
Enter
Sleep 2s

Type "git cx --provider gemini --candidates 3"
Enter
Sleep 1s

# stateSelectType: "auto" がデフォルト → "feat" へ移動
Down
Sleep 0.5s
Enter
Sleep 0.5s

# stateInputScope: "auth" を入力
Type "auth"
Sleep 0.5s
Enter
Sleep 1.5s

# stateSelectMsg: AI が候補を返す → 2番目を選択
Down
Sleep 0.5s
Enter
Sleep 0.5s

# stateSelectDetailMode: デフォルトの [Generate with AI] (index 1) → [Skip] へ
Up
Sleep 0.5s
Enter
Sleep 0.5s

# stateConfirm: y でコミット
Type "y"
Sleep 1.5s
