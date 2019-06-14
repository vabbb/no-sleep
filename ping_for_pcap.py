from pwn import *

s = ssh('root', '10.10.18.1', password='avviso', port=22)

def ls():
    sh.clean()
    sh.sendline('ls')
    files =  sh.recv(timeout=1).split()
    if len(files) > 0:
        return files
    return []

sh = s.process('/bin/sh', env={'PS1':''})
sh.sendline('cd /home/vabbb/pcaps/done')
print (sh.clean())


while True:
    files = ls()
    for file in files:
        scp = ssh('root', '10.10.18.1', password='avviso', port=22)
        a = scp.download_file('/home/vabbb/pcaps/done/'+file, './'+file)
        print (a)
        scp.close()
        sh.sendline("mv "+'/home/vabbb/pcaps/done/'+file+" /home/vabbb/archive/")
    print( "trying again in 10s...")
    time.sleep(10)

sh.interactive()