package epub

type EpubSrc interface {
	GetMenu()
	GetVolume(volume *Volume)
}
