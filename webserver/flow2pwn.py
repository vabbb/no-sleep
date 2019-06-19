def flow2pwn(flow):
    ip = flow["dstIP"]
    port = flow["dstPort"]

    script = 'from pwn import *\n'
    script += "proc = remote('{}', {})\n\n".format(ip, port)
    script += "context.log_level = 'DEBUG'\n"

    for message in flow['nodes']:
        if message['fromSrc']:
                script += 'proc.send('
                script += str(message['blob'])[1:]
                script += ')\n'

        else:
            for _ in range(len(message['blob'])):
                script += 'proc.recvuntil('
                script += str(message['blob'][-20:]).replace('\n','\\n')[1:]
                script += ')\n'
                break

    script += '\nproc.interactive()\n'

    return script
