package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type attachmentCell struct {
	widget.BaseWidget
	att *attachment

	uiface *ui
	ctrl   *genCtrl
}

func newAttachmentCell(a *attachment, u *ui, c *genCtrl) *attachmentCell {
	ret := &attachmentCell{att: a, uiface: u, ctrl: c}
	ret.ExtendBaseWidget(ret)
	return ret
}

func (m *attachmentCell) avatarResource() fyne.Resource {

	return theme.FyneLogo()
}

/*
	func (a *attachmentCell) setAttachment(new *attachment) {
		a.att = new
		a.Refresh()
	}
*/

func (a *attachmentCell) CreateRenderer() fyne.WidgetRenderer {
	name := widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	name.Wrapping = fyne.TextTruncate
	download := widget.NewButtonWithIcon("Download", theme.DownloadIcon(), func() {

		LetInteractWithDownloadAttachmentDialog(a.att.title, a.uiface, a.ctrl)
	})
	return &attachmentRenderer{a: a,
		top:  name,
		main: download, pic: widget.NewIcon(nil), sep: widget.NewSeparator()}
}

type attachmentRenderer struct {
	a    *attachmentCell
	top  *widget.Label
	main *widget.Button
	pic  *widget.Icon
	sep  *widget.Separator
}

func (a *attachmentRenderer) Destroy() {
}

func (a *attachmentRenderer) Layout(s fyne.Size) {
	remainWidth := s.Width - iconSize - theme.Padding()*2
	remainStart := iconSize + theme.Padding()*2

	a.pic.Resize(fyne.NewSize(iconSize, iconSize))
	a.pic.Move(fyne.NewPos(theme.Padding(), theme.Padding()))

	a.top.Move(fyne.NewPos(remainStart, -theme.Padding()))
	a.top.Resize(fyne.NewSize(remainWidth, a.top.MinSize().Height))

	a.main.Move(fyne.NewPos(remainStart, a.pic.MinSize().Height+theme.Padding()*2))
	a.main.Resize(a.main.MinSize())

	a.sep.Move(fyne.NewPos(0, s.Height-theme.SeparatorThicknessSize()))
	a.sep.Resize(fyne.NewSize(s.Width, theme.SeparatorThicknessSize()))
}

func (a *attachmentRenderer) MinSize() fyne.Size {
	s1 := a.top.MinSize()
	s2 := a.main.MinSize()
	w := fyne.Max(s1.Width, s2.Width)
	return fyne.NewSize(w+iconSize+theme.Padding()*2,
		s1.Height*2+s2.Height-theme.Padding()*4)
}

func (a *attachmentRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{a.top, a.main, a.pic, a.sep}
}

func (a *attachmentRenderer) Refresh() {

	a.top.SetText(fmt.Sprintf("%s [sent file %s]", a.a.att.username, a.a.att.title))
	go a.pic.SetResource(a.a.avatarResource())
}
