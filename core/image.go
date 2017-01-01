package core

import (
	"appengine"
	//"bytes"
	//"fmt"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"

	"appengine/urlfetch"
	"image/color"
	"net/http"
	"os"
	//"strconv"

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
	defaultSize             = 50
	black       color.Color = color.RGBA{255, 255, 255, 0}

	dpi      = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile = flag.String("fontfile", "static/Inconsolata-Regular.ttf", "filename of the ttf font")
	hinting  = flag.String("hinting", "none", "none | full")
	size     = flag.Float64("size", float64(defaultSize), "font size in points")
	spacing  = flag.Float64("spacing", 1.5, "line spacing (e.g. 2 means double spaced)")
	wonb     = flag.Bool("whiteonblack", false, "white text on a black background")
	text     = string("JOJO")
)

type Rectangle struct {
	p             image.Point
	length, width int
}

type InstagramImage struct {
	Photo      *image.RGBA
	Avatar     image.Image
	Image      image.Image
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

func (instagramImage *InstagramImage) load(c appengine.Context, url string) (image.Image, error) {
	c.Debugf("Image url: ", url)

	client := urlfetch.Client(c)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	c.Debugf("Image loaded")

	return img, nil
}

func processImage(c appengine.Context, instagramImage *InstagramImage) (*image.RGBA, error) {
	var err error

	c.Debugf(fmt.Sprintf("Image: %v", instagramImage))

	instagramImage.Image, err = instagramImage.load(c, instagramImage.DisplaySrc)
	if err != nil {
		return nil, err
	}

	instagramImage.Avatar, err = instagramImage.load(c, instagramImage.AvatarSrc)
	if err != nil {
		return nil, err
	}

	c.Debugf("result: post created")

	repostedImage, err := instagramImage.addRepost()
	if err != nil {
		return nil, err
	}

	return repostedImage, nil
}

func (instagramImage *InstagramImage) addRepost() (*image.RGBA, error) {

	//jpegImage, _, _ := image.Decode(bytes.NewReader(rawImage.Content))

	repostImageFile, err := os.Open("static/repost.png")
	if err != nil {
		return nil, err
	}
	defer repostImageFile.Close()

	repostImageSource, err := png.Decode(repostImageFile)
	if err != nil {
		return nil, err
	}

	// Draw Raw Image
	rawImage := instagramImage.Image
	rawImageBounds := rawImage.Bounds()
	rawImageMarginY := rawImageBounds.Max.Y
	m := image.NewRGBA(rawImageBounds)
	draw.Draw(m, rawImageBounds, rawImage, image.ZP, draw.Src)

	// Draw Black background
	length := defaultSize*2 + len(instagramImage.Caption)*(defaultSize/2+1)
	background := Rectangle{length: length, width: defaultSize} // Colored Mask Layer
	backgroundOffset := image.Pt(0, rawImageMarginY-background.width)
	backgroundImage := background.drawShape()
	draw.Draw(m, backgroundImage.Bounds().Add(backgroundOffset), &image.Uniform{black}, image.ZP, draw.Src)

	// Draw Repost image
	repostImage := resize.Thumbnail(uint(defaultSize), uint(defaultSize), repostImageSource, resize.Lanczos3)
	repostImageOffset := image.Pt(5, rawImageMarginY-defaultSize)
	draw.Draw(m, repostImage.Bounds().Add(repostImageOffset), repostImage, image.ZP, draw.Over)

	// Draw Avatar image
	avatarImage := resize.Thumbnail(uint(defaultSize), uint(defaultSize), instagramImage.Avatar, resize.Lanczos3)
	avatarImageOffset := image.Pt(defaultSize, rawImageMarginY-defaultSize)
	draw.Draw(m, avatarImage.Bounds().Add(avatarImageOffset), avatarImage, image.ZP, draw.Over)

	// Draw Caption
	err = addLabel(m, defaultSize*2, rawImageMarginY-defaultSize/4, instagramImage.Caption)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func addLabel(img *image.RGBA, x, y int, label string) error {

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
	c.SetSrc(image.Black)

	c.DrawString(label, point)
	return nil
}

func (r Rectangle) drawShape() *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, r.length, r.width))
}

func writeImage(w http.ResponseWriter, m *image.RGBA) error {
	if err := png.Encode(w, m); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "image/png")
	return nil
}
