pkill -f './main'
rm ./main
rm nohup.out
go build main.go
nohup ./main &