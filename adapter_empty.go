package crt

type EmptyAdapter struct{}

func NewEmptyAdapter() *EmptyAdapter {
	return &EmptyAdapter{}
}

func (e *EmptyAdapter) HandleMouseButton(button MouseButton) {

}

func (e *EmptyAdapter) HandleMouseMotion(motion MouseMotion) {

}

func (e *EmptyAdapter) HandleMouseWheel(wheel MouseWheel) {

}

func (e *EmptyAdapter) HandleKeyPress() {

}

func (e *EmptyAdapter) HandleWindowSize(size WindowSize) {

}
