from string import printable
from math import ceil

def print_buf(counter, buf, n):
    temp = [('%02x' % ord(i)) for i in buf]
    second_column = ' '.join([''.join(temp[i:i + 2]) for i in range(0, len(temp), 2)])
    third_column = ''.join([c if c in printable[:-5] else '.' for c in buf])
    return '{0}: {1:<{n}}  {2}\n'.format(('%07x' % (counter * 16)), second_column, third_column, n=int(5/2*n-1))

# n -> number of bytes per column
def xxd(text, n=16):
    res = ''
    for counter, buf in [(i, text[i:i+n]) for i in range(0, len(text), n)]:
        res += print_buf(counter, buf, n)
    return res
