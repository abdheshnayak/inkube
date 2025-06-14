package nixpkghandler

import (
	"slices"
	"strings"

	fn "github.com/abdheshnayak/inkube/pkg/functions"
)

func (p *pkgHandler) rmLibFromKlFile(name string) error {
	kf, err := p.fc.GetKlFile()
	if err != nil {
		return fn.NewE(err)
	}

	np := func() []string {
		var res = make([]string, len(kf.Libraries))
		for i, v := range kf.Libraries {
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

	kf.Libraries = append(kf.Libraries[:i], kf.Libraries[i+1:]...)
	return kf.Save()
}

func (p *pkgHandler) addPackageKlFile(name string) error {
	kf, err := p.fc.GetKlFile()
	if err != nil {
		return fn.NewE(err)
	}

	if slices.Contains(kf.Packages, name) {
		return nil
	}

	kf.Packages = append(kf.Packages, name)
	return kf.Save()
}

func (p *pkgHandler) addPackageToLock(name, pkghash string) error {
	lf, err := p.fc.GetLockfile()
	if err != nil {
		return fn.NewE(err)
	}

	lf.Packages[name] = pkghash

	if err = lf.Save(); err != nil {
		return fn.NewE(err)
	}

	return nil
}
