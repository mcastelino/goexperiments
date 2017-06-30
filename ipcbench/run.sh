go build ipc_notifier/notifier.go
go build ipc_listener/listener.go
rm -f ./*.PNG
rm -f /tmp/ipc_test_uri

./listener --network="unix" --notifications="$1" &
./notifier --network="unix" --notifications="$1" --chart=unix.PNG
rm -f /tmp/ipc_test_uri

./listener --network="unixpacket" --notifications="$1" &
./notifier --network="unixpacket" --notifications="$1" --chart=unixpacket.PNG
rm -f /tmp/ipc_test_uri

#./listener --network="unixgram" --notifications="$1" &
#./notifier --network="unixgram" --notifications="$1" --chart=unixgram.PNG

./listener --uri ":8000" --network="tcp4" --notifications="$1" &
./notifier --uri ":8000" --network="tcp4" --notifications="$1" --chart=tcp4.PNG

./listener --uri ":8006" --network="tcp6" --notifications="$1" &
./notifier --uri ":8006" --network="tcp6" --notifications="$1" --chart=tcp6.PNG

#Assuming you have ImageMagick installed
display unix.PNG &
display unixpacket.PNG &
display tcp4.PNG &
display tcp6.PNG &
