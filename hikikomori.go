package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/nstratos/go-myanimelist/mal"
	"github.com/pierrre/mangadownloader"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

type Mangas struct {
	Mangaseries map[string]Detail `json:"mangaseries"`
}
type Detail struct {
	Name                  string `json:"name"`
	Dir                   string `json:"dir, omitempty"`
	Url                   string `json:"url, omitempty"`
	Lastchapterdownloaded int    `json:"lastchapterdownloaded"`
	Lastchapterread       int    `json:"lastchapterread"`
}

type byName []os.FileInfo

func (f byName) Len() int { return len(f) }
func (f byName) Less(i, j int) bool {

	chapter1 := strings.Split(f[i].Name(), ".")
	chapter2 := strings.Split(f[j].Name(), ".")
	return chapter1[0] < chapter2[0]
}
func (f byName) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

func main() {

	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetConfigName("config") // name of config file (without extension)
	err := viper.ReadInConfig()   // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	file2, err := ioutil.ReadFile(viper.GetString("yaml"))
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error mangas file: %s \n", err))
	}
	mangas := &Mangas{}
	file, _ := yaml.YAMLToJSON(file2)
	json.Unmarshal(file, &mangas)

	var syncMalCmd = &cobra.Command{
		Use:   "syncMal",
		Short: "Sync MyAnimelist manga list with config file",
		Long: `Sync MyAnimelist manga list to config file in Yaml.
            It will make the difference between what alreay there.`,
		Run: func(cmd *cobra.Command, args []string) {
			c := mal.NewClient()
			c.SetCredentials(viper.GetString("myanimelist.login"), viper.GetString("myanimelist.password"))
			c.SetUserAgent(viper.GetString("myanimelist.apikey"))
			list, _, _ := c.Manga.List("Noriak")
			for _, manga := range list.Manga {
				if manga.MyStatus == 1 {
					mangaName := manga.SeriesTitle
					mangaId := manga.SeriesMangaDBID
					mangaIdString := strconv.Itoa(mangaId)

					malLastChapter := manga.MyReadChapters
					lastChapterRead := malLastChapter

					_, ok := mangas.Mangaseries[mangaIdString]
					mangaYaml := &Detail{}
					mangaYaml.Name = mangaName
					mangaYaml.Dir = mangaName
					mangaYaml.Lastchapterdownloaded = 0
					mangaYaml.Url = ""

					if ok {
						yamlLastChapterDownloaded := mangas.Mangaseries[mangaIdString].Lastchapterdownloaded
						mangaYaml.Lastchapterdownloaded = yamlLastChapterDownloaded
						yamlUrl := mangas.Mangaseries[mangaIdString].Url
						mangaYaml.Url = yamlUrl
						yamlDir := mangas.Mangaseries[mangaIdString].Dir
						if yamlDir != "" {
							mangaYaml.Dir = yamlDir
						}
						yamlLastChapterRead := mangas.Mangaseries[mangaIdString].Lastchapterread
						if yamlLastChapterRead > malLastChapter {
							//update mal
							lastChapterRead = yamlLastChapterRead
							_, err := c.Manga.Update(mangaId, mal.MangaEntry{Status: "reading", Chapter: lastChapterRead})
							if err != nil {
								fmt.Println("Error update chapter read to myanimelist")
							} else {
								fmt.Println(mangaName + " updated")
							}
						}
					}
					mangaYaml.Lastchapterread = lastChapterRead
					mangas.Mangaseries[mangaIdString] = *mangaYaml
				}
			}
			//Savetofile
			b, _ := json.Marshal(mangas)
			y, _ := yaml.JSONToYAML(b)
			ioutil.WriteFile(viper.GetString("yaml"), y, 0777)
		},
	}

	var downloadCmd = &cobra.Command{
		Use:   "download",
		Short: "Download new chapters",
		Long: `Try to see if there is new chapters and download them.
            It will possible to read them after.`,
		Run: func(cmd *cobra.Command, args []string) {

			// Var for mangadownloader
			//outFlag := flag.String("out", "", "Output directory")
			cbzFlag := flag.Bool("cbz", true, "CBZ")
			pageDigitCountFlag := flag.Int("pagedigitcount", 4, "Page digit count")
			httpRetryFlag := flag.Int("httpretry", 10, "Http retry")
			parallelChapterFlag := flag.Int("parallelchapter", 2, "Parallel chapter")
			parallelPageFlag := flag.Int("parallelpage", 4, "Parallel page")
			flag.Parse()
			//out := *outFlag

			options := &mangadownloader.Options{
				Cbz:             *cbzFlag,
				PageDigitCount:  *pageDigitCountFlag,
				ParallelChapter: *parallelChapterFlag,
				ParallelPage:    *parallelPageFlag,
			}

			md := mangadownloader.CreateDefaultMangeDownloader()
			md.HttpRetry = *httpRetryFlag

			for key, value := range mangas.Mangaseries {
				mangaName := value.Name
				mangaDir := value.Dir
				lastChapterDownloaded := value.Lastchapterdownloaded
				urlManga := value.Url
				if urlManga != "" {
					fmt.Println("Downloading " + mangaName)
					u, err := url.Parse(urlManga)
					if err != nil {
						panic(err)
					}
					o, err := md.Identify(u)
					if err != nil {
						panic(err)
					}
					dirOut := viper.GetString("downloadDir") + mangaDir + "/"
					_ = md.DownloadManga(o.(*mangadownloader.Manga), "./mangas/", options)
					files, _ := ioutil.ReadDir(dirOut)
					var cbzFiles []os.FileInfo
					for _, f := range files {
						if strings.Contains(f.Name(), ".cbz") {
							cbzFiles = append(cbzFiles, f)
						}
					}
					sort.Sort(byName(cbzFiles))
					newChapterCbz := ""
					for _, f := range cbzFiles {
						newChapterCbz = f.Name()
					}
					newChapter := strings.Split(newChapterCbz, ".")
					newChapterInt, _ := strconv.Atoi(newChapter[0])
					if newChapterInt > lastChapterDownloaded {
						lastChapterDownloaded = newChapterInt
					}
					value.Lastchapterdownloaded = lastChapterDownloaded
					mangas.Mangaseries[key] = value
					//Savetofile
					b, _ := json.Marshal(mangas)
					y, _ := yaml.JSONToYAML(b)
					ioutil.WriteFile(viper.GetString("yaml"), y, 0777)
					fmt.Println("Finish downloading " + mangaName)
				}
			}
		},
	}

	var readCmd = &cobra.Command{
		Use:   "read",
		Short: "Read new downloaded chapter",
		Long: `Read new downloaded Mangas.
            It will also permit to save the new state of reading.`,
		Run: func(cmd *cobra.Command, args []string) {
			id := 1
			var listMangas []string
			for key, value := range mangas.Mangaseries {
				mangaName := value.Name
				lastChapterDownloaded := value.Lastchapterdownloaded
				lastChapterRead := value.Lastchapterread
				if lastChapterDownloaded != lastChapterRead {
					fmt.Printf("%3d | %3d/%3d | %s\n", id, lastChapterRead, lastChapterDownloaded, mangaName)
					listMangas = append(listMangas, key)
					id = id + 1
				}
			}
			fmt.Print("Enter id of Manga to read: ")
			var input int
			fmt.Scanln(&input)
			idMangaToRead := listMangas[input-1]
			manga := mangas.Mangaseries[idMangaToRead]
			mangaDir := manga.Dir
			lastChapter := manga.Lastchapterread
			lastChapter = lastChapter + 1
			lastChapterStr := fmt.Sprintf("%03d", lastChapter)
			linkChapter := viper.GetString("downloadDir") + mangaDir + "/" + lastChapterStr + ".cbz"
			//Exec Mcomix
			cmdMcomix := exec.Command("mcomix", "-w", linkChapter)
			if err := cmdMcomix.Start(); err != nil {
				panic(err)
			}
			fmt.Print("Enter last chapter read: ")
			var newLastChapter int
			fmt.Scanln(&newLastChapter)
			manga.Lastchapterread = newLastChapter
			mangas.Mangaseries[idMangaToRead] = manga
			//Savetofile
			b, _ := json.Marshal(mangas)
			y, _ := yaml.JSONToYAML(b)
			ioutil.WriteFile(viper.GetString("yaml"), y, 0777)

		},
	}

	var rootCmd = &cobra.Command{Use: "hikikomori"}
	rootCmd.AddCommand(syncMalCmd)
	rootCmd.AddCommand(downloadCmd)
	rootCmd.AddCommand(readCmd)
	rootCmd.Execute()
}
