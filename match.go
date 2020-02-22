package main

func match(pat, buf string) bool {
	var pi, bi int
	for {
		if pi==len(pat) {
			return bi==len(buf)
		}
		ch:=pat[pi]
		pi++
		if ch == '*' {
			// star at the end matches.
			if pi==len(pat) {
				return true
			}
			ch=pat[pi]
			for {
				if bi==len(buf) {
					return false
				}
				if buf[bi]==ch {
					break;
				}
				bi++
			}
			continue
		}
		if bi==len(buf) {
			return false
		}
		if buf[bi]!=ch {
			return false
		}
		bi++
	}
}
