package drm

import (
	"bytes"
	"fmt"
	"unsafe"
)

func (c *Card) ModeGetResources() (*ModeResources, error) {
	var res cModeCardRes
	if err := ioctl(c.fd, ioctlModeGetResources, uintptr(unsafe.Pointer(&res))); err != nil {
		return nil, fmt.Errorf("ioctl: %w", err)
	}

	ret := ModeResources{
		MinWidth:  res.minWidth,
		MaxWidth:  res.maxWidth,
		MinHeight: res.minHeight,
		MaxHeight: res.maxHeight,
	}
	if res.countConnectors > 0 {
		ret.ConnectorIDs = make([]uint32, res.countConnectors)
		res.connectorIDPtr = uintptr(unsafe.Pointer(&ret.ConnectorIDs[0]))
	}
	if res.countCRTC > 0 {
		ret.CRTCIDs = make([]uint32, res.countCRTC)
		res.crtcIDPtr = uintptr(unsafe.Pointer(&ret.CRTCIDs[0]))
	}
	if res.countEncoders > 0 {
		ret.EncoderIDs = make([]uint32, res.countEncoders)
		res.encoderIDPtr = uintptr(unsafe.Pointer(&ret.EncoderIDs[0]))
	}
	if res.countFB > 0 {
		ret.FBIDs = make([]uint32, res.fbIDPtr)
		res.fbIDPtr = uintptr(unsafe.Pointer(&ret.FBIDs[0]))
	}
	// A race could occur here if a hotplug event happens. Need logic to fire multiple
	// times and check for consistency.
	if err := ioctl(c.fd, ioctlModeGetResources, uintptr(unsafe.Pointer(&res))); err != nil {
		return nil, fmt.Errorf("ioctl: %w", err)
	}

	return &ret, nil
}

func (c *Card) ModeGetCRTC(crtcID uint32) (*ModeCRTC, error) {
	crtc := cModeCRTC{ID: crtcID}
	if err := ioctl(c.fd, ioctlModeGetCRTC, uintptr(unsafe.Pointer(&crtc))); err != nil {
		return nil, fmt.Errorf("ioctl: %w", err)
	}
	return &ModeCRTC{
		cModeCRTC: crtc,
		Name:      cToGoString(crtc.name[:]),
	}, nil
}

func (c *Card) ModeSetCRTC(set ModeCRTC) error {
	crtc := cModeCRTC{
		setConnectorsPtr: uintptr(unsafe.Pointer(&set.SetConnectors[0])),
		countConnectors:  uint32(len(set.SetConnectors)),

		ID:        set.ID,
		FBID:      set.FBID,
		X:         set.X,
		Y:         set.Y,
		GammaSize: set.GammaSize,
		ModeValid: set.ModeValid,
		cModeInfo: cModeInfo{
			Clock:      set.Clock,
			HDisplay:   set.HDisplay,
			HSyncStart: set.HSyncStart,
			HSyncEnd:   set.HSyncEnd,
			HTotal:     set.HTotal,
			HSkew:      set.HSkew,
			VDisplay:   set.VDisplay,
			VSyncStart: set.VSyncStart,
			VSyncEnd:   set.VSyncEnd,
			VTotal:     set.VTotal,
			VScan:      set.VScan,
			VRefresh:   set.VRefresh,
			Flags:      set.Flags,
			Type:       set.Type,
		},
	}
	for i := 0; i < displayModeLen && i < len(set.Name); i++ {
		crtc.cModeInfo.name[i] = set.Name[i]
	}
	fmt.Printf("%+v\n", crtc)

	if err := ioctl(c.fd, ioctlModeSetCRTC, uintptr(unsafe.Pointer(&crtc))); err != nil {
		return fmt.Errorf("ioctl: %w", err)
	}
	return nil
}

func (c *Card) ModeGetPlane(id uint32) (*ModePlane, error) {
	plane := cModeGetPlane{ID: id}
	if err := ioctl(c.fd, ioctlModeGetPlane, uintptr(unsafe.Pointer(&plane))); err != nil {
		return nil, fmt.Errorf("ioctl: %w", err)
	}

	ret := ModePlane{cModeGetPlane: plane}
	if plane.countFormatTypes > 0 {
		ret.FormatTypes = make([]uint32, plane.countFormatTypes)
		plane.formatTypePtr = uintptr(unsafe.Pointer(&ret.FormatTypes[0]))
	}
	if err := ioctl(c.fd, ioctlModeGetPlane, uintptr(unsafe.Pointer(&plane))); err != nil {
		return nil, fmt.Errorf("ioctl: %w", err)
	}
	return &ret, nil
}

func (c *Card) ModeGetPlaneResources() (*ModePlaneResources, error) {
	res := cModeGetPlaneRes{}
	if err := ioctl(c.fd, ioctlModeGetPlaneResources, uintptr(unsafe.Pointer(&res))); err != nil {
		return nil, fmt.Errorf("ioctl: %w", err)
	}

	var ret ModePlaneResources
	if res.countPlanes > 0 {
		ret = make([]uint32, res.countPlanes)
		res.planeIDPtr = uintptr(unsafe.Pointer(&ret[0]))
	}
	if err := ioctl(c.fd, ioctlModeGetPlaneResources, uintptr(unsafe.Pointer(&res))); err != nil {
		return nil, fmt.Errorf("ioctl: %w", err)
	}
	return &ret, nil
}

func (c *Card) ModeGetEncoder(id uint32) (*ModeEncoder, error) {
	encoder := cModeGetEncoder{ID: id}
	if err := ioctl(c.fd, ioctlModeGetEncoder, uintptr(unsafe.Pointer(&encoder))); err != nil {
		return nil, fmt.Errorf("ioctl: %w", err)
	}
	return &ModeEncoder{cModeGetEncoder: encoder}, nil
}

func (c *Card) ModeGetConnector(connectorID uint32) (*ModeConnector, error) {
	conn := cModeGetConnector{ID: connectorID}
	if err := ioctl(c.fd, ioctlModeGetConnector, uintptr(unsafe.Pointer(&conn))); err != nil {
		return nil, fmt.Errorf("ioctl: %w", err)
	}

	var modes []cModeInfo
	ret := ModeConnector{
		cModeGetConnector: conn,
	}
	if conn.countEncoders > 0 {
		ret.EncoderIDs = make([]uint32, conn.countEncoders)
		conn.encodersPtr = uintptr(unsafe.Pointer(&ret.EncoderIDs[0]))
	}
	if conn.countModes > 0 {
		modes = make([]cModeInfo, conn.countModes)
		conn.modesPtr = uintptr(unsafe.Pointer(&modes[0]))
	}
	if conn.countProps > 0 {
		ret.PropIDs = make([]uint32, conn.countProps)
		ret.PropValues = make([]uint64, conn.countProps)
		conn.propsPtr = uintptr(unsafe.Pointer(&ret.PropIDs[0]))
		conn.propValuesPtr = uintptr(unsafe.Pointer(&ret.PropValues[0]))
	}
	if err := ioctl(c.fd, ioctlModeGetConnector, uintptr(unsafe.Pointer(&conn))); err != nil {
		return nil, fmt.Errorf("ioctl: %w", err)
	}

	for _, mode := range modes {
		ret.Modes = append(ret.Modes, ModeInfo{
			cModeInfo: mode,
			Name:      cToGoString(mode.name[:]),
		})
	}
	return &ret, nil
}

func (c *Card) ModeGetProperty(propID uint32) (*ModeProperty, error) {
	prop := cModeGetProperty{propID: propID}
	if err := ioctl(c.fd, ioctlModeGetProperty, uintptr(unsafe.Pointer(&prop))); err != nil {
		return nil, fmt.Errorf("ioctl: %w", err)
	}

	var enums []cModePropertyEnum
	ret := ModeProperty{
		PropID: prop.propID,
		Flags:  prop.flags,
		Name:   string(bytes.Trim(prop.name[:], "\u0000")),
	}
	if prop.countValues > 0 {
		ret.Values = make([]uint64, prop.countValues)
		prop.valuesPtr = uintptr(unsafe.Pointer(&ret.Values[0]))
	}
	if prop.countEnumBlobs > 0 {
		enums = make([]cModePropertyEnum, prop.countEnumBlobs)
		prop.enumBlobPtr = uintptr(unsafe.Pointer(&enums[0]))
	}
	if err := ioctl(c.fd, ioctlModeGetProperty, uintptr(unsafe.Pointer(&prop))); err != nil {
		return nil, fmt.Errorf("ioctl: %w", err)
	}

	for _, enum := range enums {
		ret.Enums = append(ret.Enums, ModePropertyEnum{
			Value: enum.value,
			Name:  cToGoString(enum.name[:]),
		})
	}
	return &ret, nil
}

func (c *Card) ModeConnectorSetProperty(connectorID, propID uint32, value uint64) error {
	prop := cModeConnectorSetProperty{
		value:       value,
		propID:      propID,
		connectorID: connectorID,
	}
	if err := ioctl(c.fd, ioctlModeSetProperty, uintptr(unsafe.Pointer(&prop))); err != nil {
		return fmt.Errorf("ioctl: %w", err)
	}
	return nil
}

func (c *Card) ModeObjGetProperties(id, kind uint32) (*ModeObjProperties, error) {
	prop := cModeObjGetProperties{objID: id, objType: kind}
	if err := ioctl(c.fd, ioctlModeObjGetProperties, uintptr(unsafe.Pointer(&prop))); err != nil {
		return nil, fmt.Errorf("ioctl: %w", err)
	}

	ret := ModeObjProperties{ID: prop.objID, Type: prop.objType}
	if prop.countProps > 0 {
		ret.PropIDs = make([]uint32, prop.countProps)
		ret.PropValues = make([]uint64, prop.countProps)
		prop.propsPtr = uintptr(unsafe.Pointer(&ret.PropIDs[0]))
		prop.propValuesPtr = uintptr(unsafe.Pointer(&ret.PropValues[0]))
	}
	if err := ioctl(c.fd, ioctlModeObjGetProperties, uintptr(unsafe.Pointer(&prop))); err != nil {
		return nil, fmt.Errorf("ioctl: %w", err)
	}
	return &ret, nil
}

func (c *Card) ModeGetBlob(id uint32) (*ModeBlob, error) {
	blob := cModeGetBlob{blobID: id}
	if err := ioctl(c.fd, ioctlModeGetPropBlob, uintptr(unsafe.Pointer(&blob))); err != nil {
		return nil, fmt.Errorf("ioctl: %w", err)
	}

	ret := ModeBlob{ID: blob.blobID}
	if blob.length > 0 {
		ret.Data = make([]uint8, blob.length)
		blob.data = uintptr(unsafe.Pointer(&ret.Data[0]))
	}
	if err := ioctl(c.fd, ioctlModeGetPropBlob, uintptr(unsafe.Pointer(&blob))); err != nil {
		return nil, fmt.Errorf("ioctl: %w", err)
	}
	return &ret, nil
}
