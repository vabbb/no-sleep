from configuration import vm_ip

def flow2pwn(flow):
    ip = flow["dstIP"]
    port = flow["dstPort"]

    script = """from pwn import *
proc = remote('{}', {})
""".format(ip, port)

    for message in flow['nodes']:
        if message['fromSrc']:
            script += """proc.writeline("{}")\n""".format(message['printableData'][:-1])

        else:
            for _ in range(len(message['printableData'])):
                script += """proc.recvuntil("{}")\n""".format(message['printableData'][-20:].replace("\n","\\n"))
                break

    return script