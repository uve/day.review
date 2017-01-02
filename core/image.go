package core

import (
	"appengine"
	"bytes"
	//"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"

	"appengine/urlfetch"
	"image/color"
	"net/http"
	"os"
	"strconv"

	"flag"
	"fmt"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/nfnt/resize"
	//"golang.org/x/image/font"

	//"golang.org/x/image/font"
	//"golang.org/x/image/font/basicfont"
	//"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

var (
	defaultSize             = float64(50)
	black       color.Color = color.RGBA{0, 0, 0, 128}

	dpi        = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile   = flag.String("fontfile", "static/Inconsolata-Regular.ttf", "filename of the ttf font")
	hinting    = flag.String("hinting", "none", "none | full")
	size       = flag.Float64("size", float64(defaultSize), "font size in points")
	spacing    = flag.Float64("spacing", 1.5, "line spacing (e.g. 2 means double spaced)")
	wonb       = flag.Bool("whiteonblack", false, "white text on a black background")
	text       = string("JOJO")
	textMargin = float64(2.1)
)

type circle struct {
	p image.Point
	r int
}

type Rectangle struct {
	p             image.Point
	length, width int
}

type InstagramImage struct {
	Photo      *image.Image
	Avatar     *image.Image
	Image      *image.Image
	AvatarSrc  string
	DisplaySrc string
	Caption    string
	Body       string
}

func newInstagramImage(DisplaySrc, AvatarSrc, Caption string) *InstagramImage {
	instagramImage := new(InstagramImage)
	instagramImage.DisplaySrc = DisplaySrc
	instagramImage.AvatarSrc = AvatarSrc
	instagramImage.Caption = Caption

	return instagramImage
}

func (instagramImage *InstagramImage) load(c appengine.Context, url string) (*image.Image, error) {
	c.Debugf("Image url: ", url)

	client := urlfetch.Client(c)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	c.Debugf("Image loaded")

	img, err := jpeg.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	c.Debugf("Image decoded")

	return &img, nil
}

func (instagramImage *InstagramImage) processImage(c appengine.Context) error {
	var err error

	c.Debugf(fmt.Sprintf("Image: %v", instagramImage))

	instagramImage.Image, err = instagramImage.load(c, instagramImage.DisplaySrc)
	if err != nil {
		return err
	}

	instagramImage.Avatar, err = instagramImage.load(c, instagramImage.AvatarSrc)
	if err != nil {
		return err
	}

	instagramImage.Photo, err = instagramImage.addRepost(c)
	if err != nil {
		return err
	}

	return nil
}

func (instagramImage *InstagramImage) addRepost(c appengine.Context) (*image.Image, error) {
	c.Debugf("Loading repost image")

	repostImageFile, err := os.Open("static/repost.png")
	if err != nil {
		return nil, err
	}
	defer repostImageFile.Close()

	c.Debugf("Repost image loaded")

	repostImageSource, err := png.Decode(repostImageFile)
	if err != nil {
		return nil, err
	}

	c.Debugf("Repost image decoded")

	c.Debugf("Draw Raw image")
	rawImage := *instagramImage.Image
	rawImageBounds := rawImage.Bounds()
	rawImageMarginY := float64(rawImageBounds.Max.Y)
	m := image.NewRGBA(rawImageBounds)
	draw.Draw(m, rawImageBounds, rawImage, image.ZP, draw.Src)

	c.Debugf("Draw Black background")
	length := defaultSize*textMargin + float64(len(instagramImage.Caption))*(defaultSize/2+1)
	background := Rectangle{length: int(length), width: int(defaultSize)} // Colored Mask Layer
	backgroundOffset := image.Pt(0, int(rawImageMarginY-defaultSize))
	backgroundImage := background.drawShape()
	draw.Draw(m, backgroundImage.Bounds().Add(backgroundOffset), &image.Uniform{black}, image.ZP, draw.Over)

	c.Debugf("Draw Repost image")
	repostImage := resize.Thumbnail(uint(defaultSize*1.0), uint(defaultSize*1.0), repostImageSource, resize.Lanczos3)
	repostImageOffset := image.Pt(0, int(rawImageMarginY-defaultSize*1.0))
	draw.Draw(m, repostImage.Bounds().Add(repostImageOffset), repostImage, image.ZP, draw.Over)

	c.Debugf("Draw Avatar image")
	avatarImage := resize.Thumbnail(uint(defaultSize), uint(defaultSize), *instagramImage.Avatar, resize.Lanczos3)
	avatarImageOffset := image.Pt(int(defaultSize), int(rawImageMarginY-defaultSize))
	avatarImageCenter := image.Pt(int(defaultSize/2), int(defaultSize/2))
	avatarCircle := &circle{avatarImageCenter, int(float64(defaultSize) / 2.1)}
	//draw.Draw(m, avatarImage.Bounds().Add(avatarImageOffset), avatarImage, image.ZP, draw.Over)
	draw.DrawMask(m, avatarImage.Bounds().Add(avatarImageOffset), avatarImage, image.ZP, avatarCircle, image.ZP, draw.Over)

	c.Debugf("Draw Draw Caption")
	err = addLabel(m, float64(defaultSize)*textMargin, float64(rawImageMarginY)-float64(defaultSize)/4, instagramImage.Caption)
	if err != nil {
		return nil, err
	}

	c.Debugf("jpeg encode...")
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, m, &jpeg.Options{Quality: 70})
	if err != nil {
		return nil, err
	}

	c.Debugf("jpeg encoded, size: ", len(buf.Bytes()))

	result, err := jpeg.Decode(buf)
	if err != nil {
		return nil, err
	}

	c.Debugf("jpeg decoded, size: ", len(buf.Bytes()))

	return &result, nil
}

func (c *circle) ColorModel() color.Model {
	return color.AlphaModel
}

func (c *circle) Bounds() image.Rectangle {
	return image.Rect(c.p.X-c.r, c.p.Y-c.r, c.p.X+c.r, c.p.Y+c.r)
}

func (c *circle) At(x, y int) color.Color {
	xx, yy, rr := float64(x-c.p.X)+0.5, float64(y-c.p.Y)+0.5, float64(c.r)
	if xx*xx+yy*yy < rr*rr {
		return color.Alpha{255}
	}
	return color.Alpha{0}
}

func addLabel(img *image.RGBA, x, y float64, label string) error {

	b, err := ioutil.ReadFile(*fontfile)
	if err != nil {
		return err
	}
	f, err := truetype.Parse(b)
	if err != nil {
		return err
	}

	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}

	c := freetype.NewContext()
	c.SetDPI(*dpi)
	c.SetFont(f)
	c.SetFontSize(*size)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.White)

	c.DrawString(label, point)
	return nil
}

func (r Rectangle) drawShape() *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, r.length, r.width))
}

func writeImage(w http.ResponseWriter, m *image.Image) error {
	var buffer bytes.Buffer
	err := jpeg.Encode(&buffer, *m, &jpeg.Options{Quality: 70})
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		return err
	}
	return nil
}
