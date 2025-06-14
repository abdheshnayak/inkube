package nixpkghandler

import (
	"slices"
	"strings"

	fn "github.com/abdheshnayak/inkube/pkg/functions"
)

func (p *pkgHandler) rmPackageFromKlFile(name string) error {
	kf, err := p.fc.GetKlFile()
	if err != nil {
		return fn.NewE(err)
	}

	np := func() []string {
		var res = make([]string, len(kf.Packages))
		for i, v := range kf.Packages {
			s := strings.Split(v, "@")
			if len(s) >= 1 {
				res[i] = s[0]
			}
		}

		return res
	}()

	i := slices.Index(np, name)
	if i == -1 {
		return nil
	}

	kf.Packages = append(kf.Packages[:i], kf.Packages[i+1:]...)
	return kf.Save()
}

func (p *pkgHandler) addLibToKlFile(name string) error {
	kf, err := p.fc.GetKlFile()
	if err != nil {
		return fn.NewE(err)
	}

	if slices.Contains(kf.Libraries, name) {
		return nil
	}

	kf.Libraries = append(kf.Libraries, name)
	return kf.Save()
}

func (p *pkgHandler) addLibToLock(name, pkghash string) error {
	lf, err := p.fc.GetLockfile()
	if err != nil {
		return fn.NewE(err)
	}

	lf.Libraries[name] = pkghash

	if err = lf.Save(); err != nil {
		return fn.NewE(err)
	}

	return nil
}
