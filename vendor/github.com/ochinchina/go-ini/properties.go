package ini

type Properties struct {
	ini *Ini
}

func NewProperties() *Properties {
	return &Properties{ini: NewIni()}
}

func (p *Properties) Load(sources ...interface{}) {
	p.ini.Load(sources)
}

func (p *Properties) GetProperty(key string) (string, error) {
	return p.ini.GetValue(p.ini.GetDefaultSectionName(), key)
}

func (p *Properties) GetPropertyWithDefault(key string, defValue string) string {
	v, err := p.GetProperty(key)
	if err == nil {
		return v
	}
	return defValue
}

func (p *Properties) GetBool(key string) (bool, error) {
	return p.ini.GetBool(p.ini.GetDefaultSectionName(), key)
}

func (p *Properties) GetBoolWithDefault(key string, defValue bool) bool{
	v, err := p.GetBool(key)
	if err == nil {
		return v
	} else {
		return defValue
	}
}

func (p *Properties) GetInt(key string) (int, error) {
	return p.ini.GetInt(p.ini.GetDefaultSectionName(), key)
}

func (p *Properties) GetIntWithDefault(key string, defValue int) int {
	v, err := p.GetInt(key)
	if err == nil {
		return v
	} else {
		return defValue
	}
}

func (p *Properties) GetInt64(key string) (int64, error) {
	return p.ini.GetInt64(p.ini.GetDefaultSectionName(), key)
}

func (p *Properties) GetInt64WithDefault(key string, defValue int64) int64 {
	v, err := p.GetInt64(key)
	if err == nil {
		return v
	} else {
		return defValue
	}
}

func (p *Properties) GetUint64(key string) (uint64, error) {
	return p.ini.GetUint64(p.ini.GetDefaultSectionName(), key)
}

func (p *Properties) GetUint64WithDefault(key string, defValue uint64) uint64 {
	v, err := p.GetUint64(key)
	if err == nil {
		return v
	} else {
		return defValue
	}
}

func (p *Properties) GetUint(key string) (uint, error) {
	return p.ini.GetUint(p.ini.GetDefaultSectionName(), key)
}

func (p *Properties) GetUintWithDefault(key string, defValue uint) uint {
	v, err := p.GetUint(key)
	if err == nil {
		return v
	} else {
		return defValue
	}
}

func (p *Properties) GetFloat32(key string) (float32, error) {
	return p.ini.GetFloat32(p.ini.GetDefaultSectionName(), key)
}

func (p *Properties) GetFloat32WithDefault(key string, defValue float32) float32 {
	v, err := p.GetFloat32(key)
	if err == nil {
		return v
	} else {
		return defValue
	}
}

func (p *Properties) GetFloat64(key string) (float64, error) {
	return p.ini.GetFloat64(p.ini.GetDefaultSectionName(), key)
}

func (p *Properties) GetFloat64WithDefault(key string, defValue float64) float64 {
	v, err := p.GetFloat64(key)
	if err == nil {
		return v
	} else {
		return defValue
	}
}
