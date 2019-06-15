#!/bin/sh
DBPATH="/opt/mongodata/"
HOST="10.10.18.1"
cd webserver
sudo python webserver.py &
cd ..
mongod --dbpath ${DBPATH}" &
cd bin
./tcp_assembler -nodebug -d pcaps -a archive -nowait &
cd ..
cd bin/pcaps
python2 ../../ping_for_pcap.py &
cd ../..
scp rm_old_and_mv_to_done.py vabbb@${HOST}:"/home/vabbb/pcaps/" &
ssh -t root@${HOST} 'cd pcaps && sudo tcpdump -i ens3 -G 20 -z ./rm_old_and_mv_to_done.py -w dump_%Y-%m-%d_%H:%M:%S.pcap tcp port 8080 or 443 or 80 or 9876 or not 22' &

