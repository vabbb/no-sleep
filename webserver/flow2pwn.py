import configuration as c

def flow2pwn(connection):
    IP = connection['srcIP']
    port = connection['srcPort']

    script = """from pwn import *

p = remote('{}', {})\n\n""".format(IP, port)

    for message in connection['flows']:
        ip_data = message['src'][:message['src'].find(':')]
        if ip_data == c.vm_ip:
            script += "p.recvuntil('{}')\n".format(message['data'])
        else:
            script += "p.sendline('{}')\n".format(message['data'][:-1])
    
    return script