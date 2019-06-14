#!/usr/bin/env python2
from sys import argv
from time import sleep
import os

COUNT = 120
DIRNAME = './done/'

os.chdir(DIRNAME)
files = os.listdir(".")

if len(files) > COUNT:
   oldest_file = sorted(files)[0]
   os.remove(oldest_file)

to_rename= '../' + argv[1]
print "to rename: ", to_rename

os.rename('../' + argv[1], argv[1])
