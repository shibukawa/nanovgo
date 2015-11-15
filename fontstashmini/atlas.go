package fontstashmini

import (
	"errors"
)

type AtlasNode struct {
	x, y, width int16
}

type Atlas struct {
	width, height int
	nodes         []AtlasNode
}

func newAtlas(width, height, nnode int) *Atlas {
	atlas := &Atlas{
		width:  width,
		height: height,
		nodes:  make([]AtlasNode, 1, nnode),
	}
	atlas.nodes[0].x = 0
	atlas.nodes[0].y = 0
	atlas.nodes[0].width = int16(width)
	return atlas
}

func (atlas *Atlas) rectFits(i, w, h int) int {
	x := int(atlas.nodes[i].x)
	y := int(atlas.nodes[i].y)
	if x+w > atlas.width {
		return -1
	}
	spaceLeft := w
	for spaceLeft > 0 {
		if i == len(atlas.nodes) {
			return -1
		}
		y = fons__maxi(y, int(atlas.nodes[i].y))
		if y+h > atlas.height {
			return -1
		}
		spaceLeft -= int(atlas.nodes[i].width)
		i++
	}
	return y
}

func (atlas *Atlas) addSkylineLevel(idx, x, y, w, h int) {
	atlas.insertNode(idx, x, y+h, w)
	for i := idx + 1; i < len(atlas.nodes); i++ {
		if atlas.nodes[i].x < atlas.nodes[i-1].x+atlas.nodes[i-1].width {
			shrink := atlas.nodes[i-1].x + atlas.nodes[i-1].width - atlas.nodes[i].x
			atlas.nodes[i].x += shrink
			atlas.nodes[i].width -= shrink
			if atlas.nodes[i].width <= 0 {
				atlas.removeNode(i)
				i--
			} else {
				break
			}
		} else {
			break
		}
	}

	for i := 0; i < len(atlas.nodes)-1; i++ {
		if atlas.nodes[i].y == atlas.nodes[i+1].y {
			atlas.nodes[i].width += atlas.nodes[i+1].width
			atlas.removeNode(i + 1)
			i--
		}
	}
}

func (atlas *Atlas) insertNode(idx, x, y, w int) {
	node := AtlasNode{
		x:     int16(x),
		y:     int16(y),
		width: int16(w),
	}
	atlas.nodes = append(atlas.nodes[:idx], append([]AtlasNode{node}, atlas.nodes[idx:]...)...)
}

func (atlas *Atlas) removeNode(idx int) {
	atlas.nodes = append(atlas.nodes[:idx], atlas.nodes[idx+1:]...)
}

func (atlas *Atlas) addRect(rw, rh int) (bestX, bestY int, err error) {
	bestH := atlas.height
	bestW := atlas.width
	bestI := -1
	bestX = -1
	bestY = -1
	for i, node := range atlas.nodes {
		y := atlas.rectFits(i, rw, rh)
		if y != -1 {
			if y+rh < bestH || ((y+rh == bestH) && (int(node.width) < bestW)) {
				bestI = i
				bestW = int(node.width)
				bestH = y + rh
				bestX = int(node.x)
				bestY = y
			}
		}
	}
	if bestI == -1 {
		err = errors.New("can't find space")
		return
	}
	// Perform the actual packing.
	atlas.addSkylineLevel(bestI, bestX, bestY, rw, rh)
	return
}

func (atlas *Atlas) reset(width, height int) {
	atlas.width = width
	atlas.height = height
	if len(atlas.nodes) != 1 {
		atlas.nodes = make([]AtlasNode, 1, cap(atlas.nodes))
		atlas.nodes[0].x = 0
		atlas.nodes[0].y = 0
		atlas.nodes[0].width = int16(width)
	}
}
