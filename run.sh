#!/bin/sh
DBPATH="/opt/mongodata/"
HOST="10.10.18.1"
$(cd webserver && konsole --noclose -e "sudo python webserver.py") &
$(konsole --noclose -e "mongod --dbpath ${DBPATH}") &
$(cd bin && konsole --noclose -e "./tcp_assembler -nodebug -d pcaps -a archive -nowait") &
$(cd bin/pcaps && konsole --noclose -e "python2 ../../ping_for_pcap.py") &
$(scp rm_old_and_mv_to_done.py vabbb@${HOST}:"/home/vabbb/pcaps/") &
$(konsole --noclose -e "ssh -t vabbb@${HOST} 'cd pcaps && sudo tcpdump -i ens3 -G 60 -z ./rm_old_and_mv_to_done.py -w dump_%Y-%m-%d_%H:%M:%S.pcap tcp port 8080 or 443 or 80 or 9876 or not 22'") &