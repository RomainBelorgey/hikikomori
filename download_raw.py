from selenium import webdriver
import sys
import os
import urllib
import yaml

def download_image(element, name):
    src = element.get_attribute('src')
    urllib.urlretrieve(src, name)


#url = sys.argv[1]
firstImage = sys.argv[1]
directory = sys.argv[2]
driver = webdriver.Firefox()
driver.implicitly_wait(30)

#driver.get("http://cartoon.media.daum.net/webtoon/view/BREAKER")
driver.get("http://cartoon.media.daum.net/webtoon/view/asura")
downloads = driver.find_elements_by_xpath("//a[@class='link_wt']")

href = []
# Get Chapters downloadable in page 1
for d in downloads:
    href.append(d.get_attribute('href'))

page = 1
# Get Chapters downloadable other pages
try:
    while True:
        page = page + 1
        element = driver.find_element_by_xpath("//a[contains(text(),'" + str(page) + "')]")
        element.click()
        downloads = driver.find_elements_by_xpath("//a[@class='link_wt']")
        for d in downloads:
            href.append(d.get_attribute('href'))
except:
    pass

# Browse chapters to get scans
for h in href:
    driver.get(h)
    if "login" not in driver.current_url:
        manga = driver.find_element_by_xpath("//dd[@class='txt_title']").text
        chapter = driver.find_element_by_xpath("//dd[@class='txt_episode']").text
        saving = directory + "/" + manga + "/" + chapter + "/"
        if not os.path.exists(saving):
            os.makedirs(saving)

            imgsave = ""
            numberPage = 1
            while True:
                if "left" in firstImage:
                    imgleft = driver.find_element_by_xpath("//div[@id='garoViewer']//div[@aria-hidden='false']/div[@class='inner_img']/img")
                    if imgsave == imgleft:
                        break
                    download_image(imgleft,saving + str(numberPage).zfill(3) + ".png")
                    numberPage = numberPage + 1
                    imgright = driver.find_element_by_xpath("//div[@id='garoViewer']//div[@aria-hidden='false']/div[@class='inner_img']/img[2]")
                    download_image(imgright,saving + str(numberPage).zfill(3) + ".png")
                    nextChapter = driver.find_element_by_xpath("//div[@id='screenView']/a[2]")
                    nextChapter.click()
                    imgsave = imgleft
                else:
                    imgright = driver.find_element_by_xpath("//div[@id='garoViewer']//div[@aria-hidden='false']/div[@class='inner_img']/img[2]")
                    if imgsave == imgright:
                        break
                    download_image(imgright,saving + str(numberPage).zfill(3) + ".png")
                    numberPage = numberPage + 1
                    imgleft = driver.find_element_by_xpath("//div[@id='garoViewer']//div[@aria-hidden='false']/div[@class='inner_img']/img")
                    download_image(imgleft,saving + str(numberPage).zfill(3) + ".png")
                    nextChapter = driver.find_element_by_xpath("//div[@id='screenView']/a")
                    nextChapter.click()
                    imgsave = imgright
                numberPage = numberPage + 1

driver.quit()
