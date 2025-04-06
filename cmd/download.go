package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// QueryToFile 将 SPARQL 查询结果写入指定文件，避免 OOM
func download(path string) error {
	// 检查文件是否已存在，如果存在则跳过下载
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("文件 %s 已存在，跳过下载\n", path)
		return nil
	}

	query := `
PREFIX dblp: <https://dblp.org/rdf/schema#>
PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#>
SELECT ?pub ?title ?page ?author ?creator ?author_name ?ordinal ?stream ?stream_name ?affiliation
WHERE {
  ?stream a dblp:Stream .
  VALUES ?stream {
    <https://dblp.org/streams/conf/chi>
<https://dblp.org/streams/conf/uist>
<https://dblp.org/streams/conf/iui>
<https://dblp.org/streams/conf/cscw>
<https://dblp.org/streams/conf/group>
<https://dblp.org/streams/conf/tabletop>
<https://dblp.org/streams/conf/automotiveUI>
<https://dblp.org/streams/conf/candc>
<https://dblp.org/streams/conf/chiplay>
<https://dblp.org/streams/conf/ci2>
<https://dblp.org/streams/conf/dev>
<https://dblp.org/streams/conf/cui>
<https://dblp.org/streams/conf/ACMdis>
<https://dblp.org/streams/conf/eics>
<https://dblp.org/streams/conf/etra>
<https://dblp.org/streams/conf/hri>
<https://dblp.org/streams/conf/icmi>
<https://dblp.org/streams/conf/acmidc>
<https://dblp.org/streams/conf/tvx>
<https://dblp.org/streams/conf/iui>
<https://dblp.org/streams/conf/mhci>
<https://dblp.org/streams/conf/recsys>
<https://dblp.org/streams/conf/sui>
<https://dblp.org/streams/conf/tei>
<https://dblp.org/streams/conf/um>
<https://dblp.org/streams/conf/huc>
<https://dblp.org/streams/conf/iswc>
<https://dblp.org/streams/conf/assets>
<https://dblp.org/streams/conf/vrst>
<https://dblp.org/streams/conf/avi>
<https://dblp.org/streams/conf/mum>
<https://dblp.org/streams/conf/mc>
<https://dblp.org/streams/conf/graphicsinterface>
<https://dblp.org/streams/conf/mm>
<https://dblp.org/streams/conf/interact>
<https://dblp.org/streams/conf/nordichi>
<https://dblp.org/streams/conf/ozchi>
<https://dblp.org/streams/conf/acii>
<https://dblp.org/streams/conf/vr>
<https://dblp.org/streams/conf/ismar>
<https://dblp.org/streams/conf/visualization>
<https://dblp.org/streams/conf/vissym>
<https://dblp.org/streams/conf/siggraph>
<https://dblp.org/streams/conf/siggrapha>
<https://dblp.org/streams/conf/sigcse>
<https://dblp.org/streams/conf/vl>
<https://dblp.org/streams/conf/lats>
<https://dblp.org/streams/journals/ijcci>
<https://dblp.org/streams/journals/tomccap>
<https://dblp.org/streams/journals/uais>
<https://dblp.org/streams/journals/tog>
<https://dblp.org/streams/journals/cgf>
<https://dblp.org/streams/journals/cg>
<https://dblp.org/streams/journals/taccess>
<https://dblp.org/streams/journals/imwut>
<https://dblp.org/streams/journals/thms>
<https://dblp.org/streams/journals/tochi>
<https://dblp.org/streams/journals/ijmms>
<https://dblp.org/streams/journals/hhci>
<https://dblp.org/streams/journals/iwc>
<https://dblp.org/streams/journals/ijhci>
<https://dblp.org/streams/journals/behaviourIT>
<https://dblp.org/streams/journals/tvcg>
<https://dblp.org/streams/journals/pacmhci>
<https://dblp.org/streams/journals/taffco>
<https://dblp.org/streams/journals/thri>
<https://dblp.org/streams/journals/vr>
<https://dblp.org/streams/journals/ijim>
  }
  ?stream rdfs:label ?stream_name .
  ?pub dblp:publishedInStream ?stream.
  ?pub dblp:title ?title .
  ?pub dblp:hasSignature ?sig .
  ?sig dblp:signatureCreator ?creator .
  ?sig dblp:signatureOrdinal ?ordinal .
  ?creator rdfs:label ?author_name .
  ?pub dblp:pagination ?page .
  OPTIONAL { ?creator dblp:primaryAffiliation ?affiliation }
}
ORDER BY ASC(?author_name)
`

	req, err := http.NewRequest("POST", "https://sparql.dblp.org/sparql", bytes.NewBufferString(query))
	if err != nil {
		return fmt.Errorf("构建请求失败: %w", err)
	}
	req.Header.Set("Accept", "text/tab-separated-values")
	req.Header.Set("Content-Type", "application/sparql-query")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	outFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer outFile.Close()

	// 获取 Content-Length（可能没有）
	contentLength, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
	var downloaded int64
	progressBarWidth := 40
	lastPrint := time.Now()

	// 包装 reader 加入进度显示
	progressReader := io.TeeReader(resp.Body, &progressWriter{
		Total:       int64(contentLength),
		Downloaded:  &downloaded,
		LastPrinted: &lastPrint,
		Width:       progressBarWidth,
	})

	_, err = io.Copy(outFile, progressReader)
	if err != nil {
		return fmt.Errorf("写入失败: %w", err)
	}

	fmt.Println("\n✅ 下载完成！")
	return nil
}

// 进度条写入器
type progressWriter struct {
	Total       int64
	Downloaded  *int64
	LastPrinted *time.Time
	Width       int
}

func (p *progressWriter) Write(data []byte) (int, error) {
	n := len(data)
	*p.Downloaded += int64(n)

	// 每 300ms 刷新一次
	if time.Since(*p.LastPrinted) > 300*time.Millisecond {
		*p.LastPrinted = time.Now()
		if p.Total > 0 {
			percent := float64(*p.Downloaded) / float64(p.Total)
			bar := int(percent * float64(p.Width))
			fmt.Printf("\r[%s%s] %.1f%%",
				strings.Repeat("█", bar),
				strings.Repeat(" ", p.Width-bar),
				percent*100)
		} else {
			fmt.Printf("\r已下载: %d KB", *p.Downloaded/1024)
		}
	}
	return n, nil
}
