import hashlib
from collections import defaultdict

def hashit(s):
    m = hashlib.md5(s.encode())
    return m.hexdigest()


dd = defaultdict(int)
f = open("../rockyou-samples.md5.txt")
for line in f.readlines():
    dd[line.strip()] += 1


def it():
    needed = len(dd.keys())
    print("gogogogo")
    cs = "0123456789abcdefghijklmnopqrstuvwxyz"
    hm = {}
    for a in cs:
        for b in cs:
            for c in cs:
                for d in cs:
                    for e in cs:
                        pw = "".join([a,b,c,d,e])
                        hw = hashit(pw)
                        if hw in dd and hw not in hm:
                            needed-=1
                            hm[hw] = pw

    print("OK")
    
    wf = open("twat.txt", "w")
    wf.writelines(["{},{}\n".format(dd[h],hm[h]) for h in hm.keys()])
    wf.close()
    f.close()
    return
it()
