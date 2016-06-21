#!/usr/bin/python3

import sys
import os


def checkFile(filepath):
    with open(filepath, 'r') as f:
        lines = f.readlines()
        offendingLines = [
            (i, l)
            for i, l
            in zip(range(len(lines)), lines)
            if len(l) > 89
        ]
        if (len(offendingLines) > 0):
            print(
                    "{} contains {} lines over 89 chars long".format(
                        filepath,
                        len(offendingLines)
                    )
                 )
            for i, l in offendingLines:
                print("{} ({}): {}".format(i, len(l), l))
        return len(offendingLines)

totalOffending = 0
for f in sys.argv:
    totalOffending += checkFile(f)

if(totalOffending > 0):
    print("{} total offending lines".format(totalOffending))
