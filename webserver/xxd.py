from string import printable
from math import ceil

def print_buf(counter, buf):
    temp = [('%02x' % ord(i)) for i in buf]
    second_column = ' '.join([''.join(temp[i:i + 2]) for i in range(0, len(temp), 2)])
    third_column = ''.join([c if c in printable[:-5] else '.' for c in buf])
    return '{0}: {1:<39}  {2}\n'.format(('%07x' % (counter * 16)), second_column, third_column)

def xxd(text):
    res = ''
    for counter, buf in [(i, text[i:i+16]) for i in range(0, len(text), 16)]:
        res += print_buf(counter, buf)
    return res
