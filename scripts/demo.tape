Output docs/tdocker.gif

Require tdocker

Set Shell bash
Set FontSize 16
Set Width 1200          # ~120 cols so expanded Ports/Network detail values fit
Set Height 520          # ~22 rows
Set Padding 20
Set TypingSpeed 70ms

Hide
Type "clear"
Enter
Show

Type "tdocker"
Sleep 700ms
Enter
Sleep 4.2s              # startup + dwell on the grouped list

Left                    # collapse Compose group -> "▸ shop (3 running)"
Sleep 1.5s
Right                   # expand it again
Sleep 1.2s

Type "j"                # -> shop/db
Sleep 600ms
Right                   # expand inline details (async inspect, then ports + network)
Sleep 3.8s
Left                    # collapse details
Sleep 800ms

Type "j"                # -> shop/web (nginx)
Sleep 800ms
Type "l"                # logs (access-log lines)
Sleep 2.4s
Type "q"                # back
Sleep 700ms

Type "e"                # exec a shell into the container
Sleep 1.9s
Type "ls /"
Enter
Sleep 1.4s
Type "exit"
Enter
Sleep 1.1s

Type "c"                # copy container ID (toast)
Sleep 1.6s
Type "q"                # quit
Sleep 1.5s
