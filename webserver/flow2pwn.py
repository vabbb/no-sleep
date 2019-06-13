import configuration as c
from pprint import pprint

def flow2pwn(connection):
    dst = connection[0]['dst']
    IP, port = dst.split(':')

    script = """from pwn import *

p = remote('{}', {})\n\n""".format(IP, port)

    for message in connection:
        ip_data = message['src'][:message['src'].find(':')]
        if ip_data == c.vm_ip:
            script += "p.recvuntil('{}')\n".format(message['data'])
        else:
            script += "p.sendline('{}')\n".format(message['data'][:-1])
    
    return script