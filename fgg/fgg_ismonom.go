package fgg

import (
	"fmt"
	"reflect"
	"strings"
)

var _ = fmt.Errorf
var _ = reflect.Append
var _ = strings.Compare

// Return true if *not* nomono
func IsMonomOK(p FGGProgram) (bool, string) {
	ds := p.GetDecls()
	for _, v := range ds {
		if md, ok := v.(MethDecl); ok {
			omega1 := Nomega{make(map[string]Type), make(map[string]MethInstan2)}
			delta := md.Psi_recv.ToDelta()
			for _, v := range md.Psi_meth.tFormals {
				delta[v.name] = v.u_I
			}
			gamma := make(Gamma)
			psi_recv := make(SmallPsi, len(md.Psi_recv.tFormals))
			for i, v := range md.Psi_recv.tFormals {
				psi_recv[i] = v.name
			}
			//psi_recv = md.Psi_recv.Hat()
			u_recv := TNamed{md.t_recv, psi_recv}
			gamma[md.x_recv] = u_recv
			omega1.us[toKey_Wt(u_recv)] = u_recv
			for _, v := range md.pDecls { // TODO: factor out
				gamma[v.name] = v.u
			}
			collectExpr2(ds, delta, gamma, md.e_body, omega1)
			if ok, msg := nomonoOmega(ds, delta, md, omega1); ok {
				return false, msg
			}
		}
	}
	return true, ""
}

// Return true if nomono
func nomonoOmega(ds []Decl, delta Delta, md MethDecl, omega1 Nomega) (bool, string) {
	for auxG2(ds, delta, omega1) {
		for _, v := range omega1.ms {
			if !isStructType(ds, v.u_recv) {
				continue
			}
			u_S := v.u_recv.(TNamed)
			if u_S.t_name == md.t_recv && v.meth == md.name {
				if occurs(md.Psi_recv, u_S.u_args) {
					return true, md.t_recv + md.Psi_recv.String() + " ->* " + md.t_recv +
						"(" + SmallPsi(u_S.u_args).String() + ")"
				}
				if occurs(md.Psi_meth, v.psi) {
					return true, md.t_recv + md.Psi_recv.String() + "." + md.name +
						md.Psi_meth.String() + " ->* " + md.name + "(" + v.psi.String() + ")"
				}
			}
		}
	}
	return false, ""
}

// Pre: len(Psi) == len(psi)
func occurs(Psi BigPsi, psi SmallPsi) bool {
	for i, v := range Psi.tFormals {
		if cast, ok := psi[i].(TNamed); ok { // !!! simplified
			for _, x := range fv(cast) {
				if x.Equals(v.name) {
					return true
				}
			}
		}
	}
	return false
}

func fv(u Type) []TParam {
	if cast, ok := u.(TParam); ok {
		return []TParam{cast}
	}
	res := []TParam{}
	cast := u.(TNamed)
	for _, v := range cast.u_args {
		res = append(res, fv(v)...)
	}
	return res
}

/* Duplication of Omega for non-ground types -- if only Go had generics! */

type Nomega struct {
	us map[string]Type
	ms map[string]MethInstan2
}

func (w Nomega) clone() Nomega {
	us := make(map[string]Type)
	ms := make(map[string]MethInstan2)
	for k, v := range w.us {
		us[k] = v
	}
	for k, v := range w.ms {
		ms[k] = v
	}
	return Nomega{us, ms}
}

func (w Nomega) Println() {
	fmt.Println("=== Type instances:")
	for _, v := range w.us {
		fmt.Println(v)
	}
	fmt.Println("--- Method instances:")
	for _, v := range w.ms {
		fmt.Println(v.u_recv, v.meth, v.psi)
	}
	fmt.Println("===")
}

// TODO: rename/refactor
type MethInstan2 struct {
	u_recv Type
	meth   Name
	psi    SmallPsi
}

func toKey_Wt2(u Type) string {
	return u.String()
}

func toKey_Wm2(x MethInstan2) string {
	return x.u_recv.String() + "_" + x.meth + "_" + x.psi.String()
}

// TODO: rename/refactor (cf. original)
func collectExpr2(ds []Decl, delta Delta, gamma Gamma, e FGGExpr, omega Nomega) bool {
	res := false
	switch e1 := e.(type) {
	case Variable:
		return res
	case StructLit:
		for _, elem := range e1.elems {
			res = collectExpr2(ds, delta, gamma, elem, omega) || res
		}
		k := toKey_Wt2(e1.u_S)
		if _, ok := omega.us[k]; !ok {
			omega.us[k] = e1.u_S
			res = true
		}
	case Select:
		return collectExpr2(ds, delta, gamma, e1.e_S, omega)
	case Call:
		res = collectExpr2(ds, delta, gamma, e1.e_recv, omega) || res
		for _, e_arg := range e1.args {
			res = collectExpr2(ds, delta, gamma, e_arg, omega) || res
		}
		gamma1 := make(Gamma)
		for k, v := range gamma {
			gamma1[k] = v
		}
		u_recv := e1.e_recv.Typing(ds, delta, gamma1, false)
		k_t := toKey_Wt2(u_recv)
		if _, ok := omega.us[k_t]; !ok {
			omega.us[k_t] = u_recv
			res = true
		}
		m := MethInstan2{u_recv, e1.meth, e1.GetTArgs()} // CHECKME: why add u_recv separately?
		k_m := toKey_Wm2(m)
		if _, ok := omega.ms[k_m]; !ok {
			omega.ms[k_m] = m
			res = true
		}
	case Assert:
		res = collectExpr2(ds, delta, gamma, e1.e_I, omega) || res
		u := e1.u_cast
		k := toKey_Wt2(u)
		if _, ok := omega.us[k]; !ok {
			omega.us[k] = u
			res = true
		}
	case StringLit: // CHECKME
		k := toKey_Wt2(STRING_TYPE)
		if _, ok := omega.us[k]; !ok {
			omega.us[k] = STRING_TYPE
			res = true // CHECKME
		}
	case Sprintf:
		k := toKey_Wt2(STRING_TYPE)
		if _, ok := omega.us[k]; !ok {
			omega.us[k] = STRING_TYPE
			res = true
		}
		for _, arg := range e1.args {
			res = collectExpr2(ds, delta, gamma, arg, omega) || res
		}
	default:
		panic("Unknown Expr kind: " + reflect.TypeOf(e).String() + "\n\t" +
			e.String())
	}
	return res
}

/* Aux */

// Return true if omega has changed
// N.B. no closure over types occurring in bounds, or *interface decl* method sigs
//func auxG(ds []Decl, omega Omega1) bool {
func auxG2(ds []Decl, delta Delta, omega Nomega) bool {
	res := false
	res = auxF2(ds, omega) || res
	res = auxI2(ds, delta, omega) || res
	res = auxM2(ds, delta, omega) || res
	res = auxS2(ds, delta, omega) || res
	// I/face embeddings
	res = auxE12(ds, omega) || res
	res = auxE22(ds, omega) || res
	return res
}

func auxF2(ds []Decl, omega Nomega) bool {
	//func auxF(ds []Decl, delta Delta, omega Omega1) bool {
	res := false
	tmp := make(map[string]Type)
	for _, u := range omega.us {
		if !isStructType(ds, u) { //|| u.Equals(STRING_TYPE) { // CHECKME
			continue
		}
		for _, u_f := range Fields(ds, u.(TNamed)) {
			cast := u_f.u
			tmp[toKey_Wt2(cast)] = cast
		}
	}
	for k, v := range tmp {
		if _, ok := omega.us[k]; !ok {
			omega.us[k] = v
			res = true
		}
	}
	return res
}

func auxI2(ds []Decl, delta Delta, omega Nomega) bool {
	res := false
	tmp := make(map[string]MethInstan2)
	for _, m := range omega.ms {
		if !IsNamedIfaceType(ds, m.u_recv) {
			continue
		}
		for _, m1 := range omega.ms {
			if !IsNamedIfaceType(ds, m1.u_recv) {
				continue
			}
			if m1.u_recv.ImplsDelta(ds, delta, m.u_recv) {
				mm := MethInstan2{m1.u_recv, m.meth, m.psi}
				tmp[toKey_Wm2(mm)] = mm
			}
		}
	}
	for k, v := range tmp {
		if _, ok := omega.ms[k]; !ok {
			omega.ms[k] = v
			res = true
		}
	}
	return res
}

func auxM2(ds []Decl, delta Delta, omega Nomega) bool {
	res := false
	tmp := make(map[string]Type)
	for _, m := range omega.ms {
		gs := methodsDelta(ds, delta, m.u_recv)
		for _, g := range gs { // Should be only g s.t. g.meth == m.meth
			if g.meth != m.meth {
				continue
			}
			eta := MakeEta2(g.Psi, m.psi)
			for _, pd := range g.pDecls {
				u_pd := pd.u.SubsEta2(eta) // HERE: need receiver subs also? cf. map.fgg "type b Eq(b)" -- methods should be ok?
				tmp[toKey_Wt2(u_pd)] = u_pd
			}
			u_ret := g.u_ret.SubsEta2(eta)
			tmp[toKey_Wt2(u_ret)] = u_ret
		}
	}
	for k, v := range tmp {
		if _, ok := omega.us[k]; !ok {
			omega.us[k] = v
			res = true
		}
	}
	return res
}

func auxS2(ds []Decl, delta Delta, omega Nomega) bool {
	res := false
	tmp := make(map[string]MethInstan2)
	clone := omega.clone()
	for _, m := range clone.ms {
		for _, u := range clone.us {
			u_recv := bounds(delta, m.u_recv) // !!! cf. plain type param
			if !isStructType(ds, u) || !u.ImplsDelta(ds, delta, u_recv) {
				continue
			}
			u_S := u.(TNamed)
			x0, xs, e := body(ds, u_S, m.meth, m.psi)
			gamma := make(Gamma)
			gamma[x0.name] = x0.u.(TNamed)
			for _, pd := range xs {
				gamma[pd.name] = pd.u
			}
			m1 := MethInstan2{u_S, m.meth, m.psi}
			k := toKey_Wm2(m1)
			//if _, ok := omega.ms[k]; !ok { // No: initial collectExpr already adds to omega.ms
			tmp[k] = m1
			res = collectExpr2(ds, delta, gamma, e, omega) || res
			//}
		}
	}
	for k, v := range tmp {
		if _, ok := omega.ms[k]; !ok {
			omega.ms[k] = v
			res = true
		}
	}
	return res
}

// Add embedded types
func auxE12(ds []Decl, omega Nomega) bool {
	res := false
	tmp := make(map[string]TNamed)
	for _, u := range omega.us {
		if !isNamedIfaceType(ds, u) { // TODO CHECKME: type param
			continue
		}
		u_I := u.(TNamed)
		td_I := getTDecl(ds, u_I.t_name).(ITypeLit)
		eta := MakeEta2(td_I.Psi, u_I.u_args)
		for _, s := range td_I.specs {
			if u_emb, ok := s.(TNamed); ok {
				u_sub := u_emb.SubsEta2(eta).(TNamed)
				tmp[toKey_Wt2(u_sub)] = u_sub
			}
		}
	}
	for k, v := range tmp {
		if _, ok := omega.us[k]; !ok {
			omega.us[k] = v
			res = true
		}
	}
	return res
}

// Propagate method instances up to embedded supertypes
func auxE22(ds []Decl, omega Nomega) bool {
	res := false
	tmp := make(map[string]MethInstan2)
	for _, m := range omega.ms {
		if !isNamedIfaceType(ds, m.u_recv) { // TODO CHECKME: type param
			continue
		}
		u_I := m.u_recv.(TNamed)
		td_I := getTDecl(ds, u_I.t_name).(ITypeLit)
		eta := MakeEta2(td_I.Psi, u_I.u_args)
		for _, s := range td_I.specs {
			if u_emb, ok := s.(TNamed); ok {
				u_sub := u_emb.SubsEta2(eta).(TNamed)
				gs := methods(ds, u_sub)
				for _, g := range gs {
					if m.meth == g.meth {
						m_emb := MethInstan2{u_sub, m.meth, m.psi}
						tmp[toKey_Wm2(m_emb)] = m_emb
					}
				}
			}
		}
	}
	for k, v := range tmp {
		if _, ok := omega.ms[k]; !ok {
			omega.ms[k] = v
			res = true
		}
	}
	return res
}
