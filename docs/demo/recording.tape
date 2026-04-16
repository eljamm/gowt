#Output docs/demo/recording.gif
Output docs/demo/recording.mp4

Set Width 800
Set Height 800
Set Theme "Catppuccin Macchiato"

Set FontFamily "JetBrainsMono Nerd Font"
Set FontSize 16
Set CaptionFont "JetBrainsMono Nerd Font"
Set CaptionFontSize 18

Set CaptionMaxKeys 10
Set CaptionInactivityTimer 700ms
Set CaptionKeyStyle "vim"
Set CaptionBoxPadding 15

Set OverlayAlignment "top-right"

Set Shell fish

Set TypingSpeed 0.05

Type "git branch --show-current"
Enter
Sleep 3s

Type "cd $(nix run .#gwt)"
Sleep 2s
Enter
Sleep 2s

Set TypingSpeed 0.1

Type "escape to clear search"
Sleep 2s
Escape
Sleep 2s

Type "owl"
Sleep 2s
Escape
Sleep 2s

Type "let's enter normal mode"
Sleep 2s

CaptionOn

Escape
Sleep 0.5s
Escape
Sleep 0.5s

Overlay@8s "VIM motions"
Sleep 2s

Type "j"
Sleep 0.5s
Type "j"
Sleep 0.5s
Type "k"
Sleep 2s

Ctrl+U
Sleep 0.5s
Ctrl+D
Sleep 2s

Ctrl+J
Sleep 0.5s
Ctrl+K
Sleep 1s

Overlay@4s "VIM motions with count"
Sleep 2s

Type "4"
Ctrl+K
Sleep 0.5s

Type "2"
Ctrl+J
Sleep 2s

Overlay@3s "Jump to top/bottom"
Sleep 2s

Type "g"
Sleep 1s

Type "G"
Sleep 1s

Overlay@3s "Jump to specific line"
Sleep 2s

Type "42g"
Sleep 2s

Type "i"
Sleep 1s
Type "lion"
Sleep 1.5s
Enter
Sleep 2s

CaptionOff

Set TypingSpeed 0.05

Type "git branch --show-current"
Enter
Sleep 3s
