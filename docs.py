#!/usr/bin/env python3

import re

def extract(fname, initrecog, startrecog, endrecog, finrecog, infofind, numargsfind, argfind):
    global gentries

    def handleBlock(name, block):
        numargs = -1
        args = []
        info = None
        for line in block:
            if numargs == -1:
                try: numargs = numargsfind(line)
                except: pass

            if info is None:
                try: info = infofind(line)
                except: pass

            try: arg = argfind(line)
            except: arg = None
            if arg:
                tag, idx, typ = arg
                try: idx = int(idx)
                except: pass
                args.append((tag, idx, typ))

        args.sort(key=lambda tup: tup[1])
        return (name, numargs, args, info)

    entries = []
    ready = False
    with open(fname) as f:
        blockname = None
        block = None
        for line in f:
            if not ready:
                try: haveinit = initrecog(line)
                except: continue
                if not haveinit:
                    continue
                ready = True

            if block is not None:
                try: haveend = endrecog(line)
                except: haveend = False
                if haveend:
                    entries.append(handleBlock(blockname, block))
                    blockname = None
                    block = None

            try: havefin = finrecog(line)
            except: havefin = False
            if havefin:
                break

            if block is None:
                try: name = startrecog(line)
                except: name = None
                if name is not None:
                    blockname = name
                    block = [line]
            else:
                block.append(line)

    gentries.append((fname, entries))


gentries = []


extract("connections/player.go",
        lambda line: re.match(r"\s*switch msg\.Method", line),
        lambda line: re.match(r"\s*case\s+\"([^\"]*)\"", line).group(1),
        lambda line: re.match(r"\s*(case\s+|default:)", line),
        lambda line: re.match(r"\s*default:", line),
        lambda line: None,
        lambda line: int(re.search(r"expectArgs\(([^)]*)\)", line).group(1)),
        lambda line: re.search(r"(?:(\w*)(?:,\s*ok\s*)?\s*:?=\s*)?msg\.Args\[(\d+)\](?:\.\(([^)]*)\))?", line).groups())

extract("connections/playerCommands.go",
        lambda line: re.match(r"func onPlayerCommand", line),
        lambda line: re.match(r"\thandle(?:Game|Room)?Command\(\"([^\"]*)\"", line).group(1),
        lambda line: re.match(r"\t\}\)", line),
        lambda line: re.match(r"\}", line),
        lambda line: re.match(r"\thandle(Game|Room)Command\(", line).group(1),
        lambda line: int(re.match(r"\thandle(?:Game|Room)?Command\(\"[^\"]*\",\s*(\d+)", line).group(1)),
        lambda line: re.search(r"(?:(\w*)(?:,\s*ok\s*)?\s*:?=\s*)?msg\.Args\[(\d+)\](?:\.\(([^)]*)\))?", line).groups())

extract("connections/serverApi.go",
        lambda line: re.match(r"func onGameServerCommand", line),
        lambda line: re.match(r"\thandleCommand\(\"([^\"]*)\"", line).group(1),
        lambda line: re.match(r"\t\}\)", line),
        lambda line: re.match(r"\}", line),
        lambda line: None,
        lambda line: int(re.match(r"\thandleCommand\(\"[^\"]*\",\s*(\d+)", line).group(1)),
        lambda line: re.search(r"(?:(\w*)(?:,\s*ok\s*)?\s*:?=\s*)?msg\.Args\[(\d+)\](?:\.\(([^)]*)\))?", line).groups())

# print(gentries)

for fname, methods in gentries:
    title = "File '{}'".format(fname)

    print("-" * len(title))
    print(title)
    print("-" * len(title))
    for method, nargs, args, info in methods:
        line = "'{}' ({})".format(method, str(nargs) if nargs != -1 else "...")
        if info:
            line += " [{}]".format(info)
        print(line)

        if nargs >= 0:
            indices = [arg[1] for arg in args]
            if indices != list(range(nargs)):
                print("    WARNING: not all arguments detected")
        else:
            print("    NOTE: variadic")


        for tag, idx, typ in args:
            print("  {}: '{}' ({})".format(str(idx), tag, typ))

        print()

    print()
