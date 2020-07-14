package wikirel_test

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sebnyberg/wikirel"
)

func Test_PageReader(t *testing.T) {
	type result struct {
		page *wikirel.Page
		err  error
	}

	nilPage := new(wikirel.Page)

	for _, tc := range []struct {
		name  string
		input string
		want  []result
	}{
		{"empty input", "", []result{{nilPage, wikirel.ErrFailedToParse}}},
		{"invalid input", "abc123", []result{{nilPage, wikirel.ErrFailedToParse}}},
		{"download example", downloadContents, []result{
			{&accessibleComputingPage, nil},
			{&anarchismPage, nil},
			{nilPage, io.EOF},
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := strings.NewReader(tc.input)
			pageReader := wikirel.NewPageReader(r)
			for _, expected := range tc.want {
				p := new(wikirel.Page)
				err := pageReader.Read(p)
				if !cmp.Equal(expected.page, p, cmpopts.IgnoreFields(wikirel.Page{}, "Text")) {
					t.Errorf("expected page did not match result\n%v", cmp.Diff(expected.page, p))
				}
				if !cmp.Equal(err, expected.err, cmpopts.EquateErrors()) {
					t.Errorf("invalid err, expected: %v, got: %v\n", expected.err, err)
				}
			}
		})
	}
}

func Test_PageStruct(t *testing.T) {
	var p wikirel.Page
	if err := xml.Unmarshal([]byte(accessibleComputingXML), &p); err != nil {
		t.Fatalf("failed to unmarshal page: %v", err)
	}

	if !cmp.Equal(p, accessibleComputingPage) {
		t.Fatalf("failed to parse page\n%v", cmp.Diff(p, accessibleComputingPage))
	}
}

func getAnarchistWikipedia(n int) string {
	sb := strings.Builder{}
	sb.WriteString(`<mediawiki>`)
	sb.WriteString(siteInfo)
	for i := 0; i < n; i++ {
		sb.WriteString(anarchismXML)
	}
	sb.WriteString(`</mediawiki>`)
	return sb.String()
}

func Benchmark_PageReader_Read(b *testing.B) {
	anarchistWikipedia := getAnarchistWikipedia(10)
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := wikirel.NewPageReader(strings.NewReader(anarchistWikipedia))
		b.StartTimer()
		pages := make([]wikirel.Page, 0)
		for {
			var p wikirel.Page
			err := r.Read(&p)
			if err != nil {
				if err == io.EOF {
					break
				}
				b.Fatalf("unexpeted error: %v\n", err)
			}
			pages = append(pages, p)
		}
	}
}

func Test_PageIndexBlockReader(t *testing.T) {
	type result struct {
		indexBlock *wikirel.MultiStreamIndex
		err        error
	}

	for _, tc := range []struct {
		name  string
		input string
		want  []result
	}{
		{"empty input", "", []result{{nil, io.EOF}}},
		{"incomplete row", "abc123", []result{{nil, wikirel.ErrBadRecord}}},
		{
			"valid indexes",
			`1:10:A
1:11:B
1:12:C
2:13:D
2:14:E
2:15:F
3:16:G`,
			[]result{
				{&wikirel.MultiStreamIndex{1, 3}, nil},
				{&wikirel.MultiStreamIndex{2, 3}, nil},
				{&wikirel.MultiStreamIndex{3, 1}, nil},
				{nil, io.EOF},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := strings.NewReader(tc.input)
			indexReader := wikirel.NewMultiStreamIndexReader(r)
			for _, expected := range tc.want {
				got, err := indexReader.ReadIndex()
				if !cmp.Equal(expected.indexBlock, got) {
					t.Errorf("invalid index, expected / got\n%v\n", cmp.Diff(expected.indexBlock, got))
				}
				if !cmp.Equal(expected.err, err, cmpopts.EquateErrors()) {
					t.Errorf("invalid err, expected: %v, got: %v\n", expected.err, err)
				}
			}
		})
	}
}

func Test_PageIndexReader(t *testing.T) {
	type result struct {
		index *wikirel.MultiStreamIndexRow
		err   error
	}

	nilIndex := new(wikirel.MultiStreamIndexRow)

	for _, tc := range []struct {
		name  string
		input string
		want  []result
	}{
		{"empty input", "", []result{{nilIndex, io.EOF}}},
		{"incomplete row", "abc123", []result{{nilIndex, wikirel.ErrBadRecord}}},
		{
			"valid indexes",
			`1:10:A
1:11:B
1:12:C
2:13:D
2:15:F
3:16:G`,
			[]result{
				{&wikirel.MultiStreamIndexRow{1, 10, "A"}, nil},
				{&wikirel.MultiStreamIndexRow{1, 11, "B"}, nil},
				{&wikirel.MultiStreamIndexRow{1, 12, "C"}, nil},
				{&wikirel.MultiStreamIndexRow{2, 13, "D"}, nil},
				{&wikirel.MultiStreamIndexRow{2, 15, "F"}, nil},
				{&wikirel.MultiStreamIndexRow{3, 16, "G"}, nil},
				{nilIndex, io.EOF},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := strings.NewReader(tc.input)
			indexReader := wikirel.NewMultiStreamIndexReader(r)
			for _, expected := range tc.want {
				got := new(wikirel.MultiStreamIndexRow)
				err := indexReader.ReadRow(got)
				if !cmp.Equal(expected.index, got) {
					t.Errorf("invalid index, expected / got\n%v\n", cmp.Diff(expected.index, got))
				}
				if !cmp.Equal(expected.err, err, cmpopts.EquateErrors()) {
					t.Errorf("invalid err, expected: %v, got: %v\n", expected.err, err)
				}
			}
		})
	}
}

const siteInfo = `<siteinfo>
	<sitename>Wikipedia</sitename>
	<dbname>enwiki</dbname>
	<base>https://en.wikipedia.org/wiki/Main_Page</base>
	<generator>MediaWiki 1.35.0-wmf.37</generator>
	<case>first-letter</case>
	<namespaces>
		<namespace key="-2" case="first-letter">Media</namespace>
		<namespace key="2303" case="case-sensitive">Gadget definition talk</namespace>
	</namespaces>
</siteinfo>`

var accessibleComputingPage = wikirel.Page{
	Title:     "AccessibleComputing",
	ID:        10,
	Namespace: 0,
	Redirect: &wikirel.Redirect{
		Title: "Computer accessibility",
	},
	Text: `#REDIRECT [[Computer accessibility]]

	{{R from move}}
	{{R from CamelCase}}
	{{R unprintworthy}}`,
}

const accessibleComputingXML = `<page>
	<title>AccessibleComputing</title>
	<ns>0</ns>
	<id>10</id>
	<redirect title="Computer accessibility" />
	<revision>
		<id>854851586</id>
		<parentid>834079434</parentid>
		<timestamp>2018-08-14T06:47:24Z</timestamp>
		<contributor>
			<username>Godsy</username>
			<id>23257138</id>
		</contributor>
		<comment>remove from category for seeking instructions on rcats</comment>
		<model>wikitext</model>
		<format>text/x-wiki</format>
		<text bytes="94" xml:space="preserve">#REDIRECT [[Computer accessibility]]

	{{R from move}}
	{{R from CamelCase}}
	{{R unprintworthy}}</text>
		<sha1>42l0cvblwtb4nnupxm6wo000d27t6kf</sha1>
	</revision>
</page>
`

var anarchismPage = wikirel.Page{
	Title:     "Anarchism",
	Namespace: 0,
	ID:        12,
	Redirect:  nil,
	// Leaving out text
	// Text: ...
}

const anarchismXML = `<page>
	<title>Anarchism</title>
	<ns>0</ns>
	<id>12</id>
	<revision>
		<id>963604419</id>
		<parentid>963528171</parentid>
		<timestamp>2020-06-20T19:02:32Z</timestamp>
		<contributor>
			<ip>80.233.88.131</ip>
		</contributor>
		<comment>I changed 'New Romanticism' to 'Neo Romanticism' because the link very mistakenly led to 'New Romantic' pop music of 1980's Britain.</comment>
		<model>wikitext</model>
		<format>text/x-wiki</format>
		<text bytes="83050" xml:space="preserve">{{short description|Political philosophy and movement}}
	{{redirect2|Anarchist|Anarchists|other uses|Anarchists (disambiguation)}}
	{{pp-move-indef}}
	{{good article}}
	{{use dmy dates|date=March 2020}}
	{{use British English|date=January 2014}}
	{{anarchism sidebar}}
	{{basic forms of government}}
	'''Anarchism''' is a [[political philosophy]] and [[Political movement|movement]] that rejects all involuntary, coercive forms of [[hierarchy]]. It [[Radical politics|radically]] calls for the abolition of the [[State (polity)|state]] which it holds to be undesirable, unnecessary, and harmful.

	The [[timeline of anarchism]] stretches back to [[prehistory]], when humans lived in anarchistic societies long before the establishment of formal states, [[Realm|kingdom]]s or [[empire]]s. With the rise of organised hierarchical bodies, [[skepticism]] toward authority also rose, but it was not until the 19th century that a self-conscious political movement emerged. During the latter half of the 19th and the first decades of the 20th century, the anarchist movement flourished in most parts of the world and had a significant role in worker's struggles for [[emancipation]]. Various [[branches of anarchism]] formed during this period. Anarchists took part in several revolutions, most notably in the [[Spanish Civil War]], where they were crushed by the [[fascists|fascist]] forces in 1939, marking the end of the [[classical era of anarchism]]. In the last decades of the 20th century and into the 21st century, the anarchist movement has been resurgent once more.

	Anarchism employs various tactics in order to meet its ideal ends; these can be broadly separated into revolutionary and evolutionary tactics. There is significant overlap between the two, which are merely descriptive. Revolutionary tactics aim to bring down authority and state, and have taken a violent turn in the past. Evolutionary tactics aim to prefigure what an anarchist society would be like. Anarchist thought, criticism, and [[Praxis (process)|praxis]] has played a part in diverse areas of human society.

	Criticism of anarchism mainly focuses on claims of it being internally inconsistent, violent, and utopian.

	== Etymology, terminology and definition ==
	{{main|Definition of anarchism and libertarianism}}
	{{see also|Glossary of anarchism}}
	The etymological origin of ''anarchism'' is from the Ancient Greek ''anarkhia'', meaning &quot;without a ruler&quot;, composed of the prefix ''an-'' (i.e. &quot;without&quot;) and the word ''arkhos'' (i.e. &quot;leader&quot; or &quot;ruler&quot;). The suffix ''[[-ism]]'' denotes the ideological current that favours [[anarchy]].{{sfnm|1a1=Bates|1y=2017|1p=128|2a1=Long|2y=2013|2p=217}} ''Anarchism'' appears in English from 1642{{sfn|Merriam-Webster|2019|loc=&quot;Anarchism&quot;}} as ''anarchisme'' and ''anarchy'' from 1539.{{sfn|Oxford English Dictionary|2005|loc=&quot;Anarchism&quot;}} Various factions within the [[French Revolution]] labelled their opponents as ''anarchists'', although few such accused shared many views with later anarchists. Many revolutionaries of the 19th century such as [[William Godwin]] (1756–1836) and [[Wilhelm Weitling]] (1808–1871) would contribute to the anarchist doctrines of the next generation, but they did not use ''anarchist'' or ''anarchism'' in describing themselves or their beliefs.{{sfn|Joll|1964|pp=27–37}}

	The first political philosopher to call himself an ''anarchist'' ({{lang-fr|anarchiste}}) was [[Pierre-Joseph Proudhon]] (1809–1865), marking the formal birth of anarchism in the mid-19th century. Since the 1890s and beginning in France,{{sfn|Nettlau|1996|p=162}} ''[[libertarianism]]'' has often been used as a synonym for anarchism{{sfn|Guerin|1970|loc=&quot;The Basic Ideas of Anarchism&quot;}} and its use as a synonym is still common outside the United States.{{sfnm|1a1=Ward|1y=2004|1p=62|2a1=Goodway|2y=2006|2p=4|3a1=Skirda|3y=2002|3p=183|4a1=Fernández|4y=2009|4p=9}} On the other hand, some use ''libertarianism'' to refer to [[Individualist anarchism|individualistic free-market philosophy]] only, referring to [[free-market anarchism]] as ''libertarian anarchism''.{{sfn|Morris|2002|p=61}}

	While [[Anti-statism|opposition to the state]] is central to anarchist thought, defining anarchism is not an easy task as there is a lot of discussion among scholars and anarchists on the matter and various currents perceive anarchism slightly differently.{{sfn|Long|2013|p=217}} Hence, it might be true to say that anarchism is a cluster of political philosophies opposing [[authority]] and [[hierarchical organization]] (including the [[State (polity)|state]], [[Anarchism and capitalism|capitalism]], [[Anarchism and nationalism|nationalism]] and all associated [[institution]]s) in the conduct of all [[human relations]] in favour of a society based on [[voluntary association]], on [[freedom]] and on [[decentralisation]], but this definition has the same shortcomings as the definition based on etymology (which is simply a negation of a ruler), or based on anti-statism (anarchism is much more than that) or even the anti-authoritarian (which is an ''[[A priori and a posteriori|a posteriori]]'' conclusion).{{sfnm|1a1=McLaughlin|1y=2007|1pp=25–29|2a1=Long|2y=2013|2pp=217}} Nonetheless, major elements of the definition of anarchism include the following:{{sfn|McLaughlin|2007|pp=25–26}}
	# The will for a non-coercive society.
	# The rejection of the state apparatus.
	# The belief that human nature allows humans to exist in or progress toward such a non-coercive society.
	# A suggestion on how to act to pursue the ideal of anarchy.

	== History ==
	{{main|History of anarchism}}

	=== Pre-modern era ===
	[[File:Paolo Monti - Servizio fotografico (Napoli, 1969) - BEIC 6353768.jpg|thumb|upright=.7|[[Zeno of Citium]] (c. 334 – c. 262 BC), whose ''[[Republic (Zeno)|Republic]]'' inspired [[Peter Kropotkin]]{{sfn|Marshall|1993|p=70}}]]
	During the prehistoric era of mankind, an established authority did not exist. It was after the creation of towns and cities that institutions of authority were established and anarchistic ideas espoused as a reaction.{{sfn|Graham|2005|pp=xi–xiv}} Most notable precursors to anarchism in the ancient world were in [[History of China|China]] and [[Ancient Greece|Greece]]. In China, [[philosophical anarchism]] (i.e. the discussion on the legitimacy of the state) was delineated by [[Taoism|Taoist philosophers]] [[Zhuang Zhou|Zhuangzi]] and [[Lao Tzu]].{{sfnm|1a1=Coutinho|1y=2016|2a1=Marshall|2y=1993|2p=54}} Likewise, anarchic attitudes were articulated by tragedians and philosophers in Greece. [[Aeschylus]] and [[Sophocles]] used the myth of [[Antigone]] to illustrate the conflict between rules set by the state and [[autonomy|personal autonomy]]. [[Socrates]] questioned Athenian authorities constantly and insisted to the right of individual freedom of consciousness. [[Cynicism (philosophy)|Cynics]] dismissed human law (''[[Nomos (sociology)|nomos]]'') and associated authorities while trying to live according to nature (''[[physis]]''). [[Stoics]] were supportive of a society based on unofficial and friendly relations among its citizens without the presence of a state.{{sfn|Marshall|1993|pp=4, 66–73}}

	During the [[Middle Ages]], there was no anarchistic activity except some ascetic religious movements in the Islamic world or in Christian Europe. This kind of tradition later gave birth to [[religious anarchism]]. In Persia, [[Mazdak]] called for an [[Egalitarianism|egalitarian society]] and the [[abolition of monarchy]], only to be soon executed by the king.{{sfn|Marshall|1993|p=86}} In [[Basra]], religious sects preached against the state. In Europe, various sects developed anti-state and libertarian tendencies. Libertarian ideas further emerged during the [[Renaissance]] with the spread of [[Rationalism|reasoning]] and [[humanism]] through Europe. Novelists fictionalised ideal societies that were based not on coercion but voluntarism. The [[Age of Enlightenment|Enlightenment]] further pushed towards anarchism with the optimism for social progress.{{sfn|Adams|2014|pp=33–63}}

	=== Modern era ===
	During the [[French Revolution]], partisan groups such as the [[Enragés]] and the {{lang|fr|[[sans-culottes]]}} saw a turning point in the fermentation of anti-state and federalist sentiments.{{sfn|Marshall|1993|p=4}} The first anarchist currents developed throughout the 18th century—[[William Godwin]] espoused [[philosophical anarchism]] in England, morally delegitimizing the state, [[Max Stirner]]'s thinking paved the way to individualism, and [[Pierre-Joseph Proudhon]]'s theory of [[Mutualism (economic theory)|mutualism]] found fertile soil in France.{{sfn|Marshall|1993|pp=4–5}} This era of classical anarchism lasted until the end of the [[Spanish Civil War|Spanish Civil War of 1936]] and is considered the golden age of anarchism.{{sfn|Marshall|1993|pp=4–5}}

	[[File:Bakunin.png|thumb|upright|Anarchist [[Mikhail Bakunin]] opposed the [[Marxism|Marxist]] aim of [[dictatorship of the proletariat]] and allied himself with the federalists in the [[International Workingmen's Association|First International]] before his expulsion by the Marxists]]
	Drawing from mutualism, [[Mikhail Bakunin]] founded [[collectivist anarchism]] and entered the [[International Workingmen's Association]], a class worker union later known as the First International that formed in 1864 to unite diverse revolutionary currents. The International became a significant political force, with [[Karl Marx]] being a leading figure and a member of its General Council. Bakunin's faction (the [[Jura Federation]]) and Proudhon's followers (the [[Mutualism (economic theory)|mutualists]]) opposed Marxist [[state socialism]], advocating political [[abstentionism]] and small property holdings.{{sfnm|1a1=Dodson|1y=2002|1p=312|2a1=Thomas|2y=1985|2p=187|3a1=Chaliand|3a2=Blin|3y=2007|3p=116}} After bitter disputes, the Bakuninists were expelled from the International by the Marxists at the [[1872 Hague Congress]].{{sfnm|1a1=Graham|1y=2019|1pp=334–336|2a1=Marshall|2y=1993|2p=24}} Bakunin famously predicted that if revolutionaries gained power by Marxist's terms, they would end up the new tyrants of workers. After being expelled, anarchists formed the [[Anarchist St. Imier International|St. Imier International]]. Under the influence of [[Peter Kropotkin]], a Russian philosopher and scientist, [[anarcho-communism]] overlapped with collectivism.{{sfn|Marshall|1993|p=5}} Anarcho-communists, who drew inspiration from the 1871 [[Paris Commune]], advocated for free federation and for the distribution of goods according to one's needs.{{sfn|Graham|2005|p=xii}}

	At the turn of the century, anarchism had spread all over the world.{{sfn|Moya|2015|p=327}} In China, small groups of students imported the humanistic pro-science version of anarcho-communism.{{sfn|Marshall|1993|pp=519–521}} Tokyo was a hotspot for rebellious youth from countries of the far east, travelling to the Japanese capital to study.{{sfnm|1a1=Dirlik|1y=1991|1p=133|2a1=Ramnath|2y=2019|2pp=681–682}} In [[Latin America]], [[Anarchism in Argentina|Argentina]] was a stronghold for [[anarcho-syndicalism]], where it became the most prominent left-wing ideology.{{sfnm|1a1=Levy|1y=2011|1p=23|2a1=Laursen|2y=2019|2p=157|3a1=Marshall|3y=1993|3pp=504–508}} During this time, a minority of anarchists adopted tactics of revolutionary [[political violence]]. This strategy became known as [[propaganda of the deed]].{{sfn|Marshall|1993|pp=633–636}} The dismemberment of the French socialist movement into many groups, and the execution and exile of many [[Communards]] to penal colonies following the suppression of the Paris Commune, favoured individualist political expression and acts.{{sfn|Anderson|2004}} Even though many anarchists distanced themselves from these terrorist acts, infamy came upon the movement.{{sfn|Marshall|1993|pp=633–636}} [[Illegalism]] was another strategy which some anarchists adopted during this period.{{sfn|Bantman|2019|p=374}}

	[[File:Makhno group.jpg|thumb|left|[[Nestor Makhno]] with members of the anarchist [[Revolutionary Insurrectionary Army of Ukraine]]]]
	Anarchists enthusiastically participated in the [[Russian Revolution]]—despite concerns—in opposition to the [[White movement|Whites]]. However, they met harsh suppression after the [[Bolshevik government]] was stabilized. Several anarchists from Petrograd and Moscow fled to Ukraine,{{sfn|Avrich|2006|p=204}} notably leading to the [[Kronstadt rebellion]] and [[Nestor Makhno]]'s struggle in the [[Free Territory]]. With the anarchists being crushed in Russia, two new antithetical currents emerged, namely [[platformism]] and [[synthesis anarchism]]. The former sought to create a coherent group that would push for revolution while the latter were against anything that would resemble a political party. Seeing the victories of the Bolsheviks in the [[October Revolution]] and the resulting [[Russian Civil War]], many workers and activists turned to [[Communist party|communist parties]], which grew at the expense of anarchism and other socialist movements. In France and the United States, members of major syndicalist movements, the [[General Confederation of Labour (France)|General Confederation of Labour]] and [[Industrial Workers of the World]], left their organisations and joined the [[Communist International]].{{sfn|Nomad|1966|p=88}}

	In the [[Spanish Civil War]], anarchists and syndicalists ([[Confederación Nacional del Trabajo|CNT]] and [[Federación Anarquista Ibérica|FAI]]) once again allied themselves with various currents of leftists. A long tradition of [[Spanish anarchism]] led to anarchists playing a pivotal role in the war. In response to the army rebellion, an anarchist-inspired movement of peasants and workers, supported by armed militias, took control of [[Barcelona]] and of large areas of rural Spain, where they [[Collective farming|collectivised]] the land.{{sfn|Bolloten|1984|p=1107}} The [[Soviet Union]] provided some limited assistance at the beginning of the war, but the result was a bitter fight among communists and anarchists at a series of events named [[May Days]] as [[Joseph Stalin]] tried to seize control of the [[Republican faction (Spanish Civil War)|Republicans]].{{sfn|Marshall|1993|pp=xi, 466}}

	=== Post-war era ===
	[[File:Rojava Sewing Cooperative.jpg|thumb|[[Rojava]] is supporting efforts for workers to form cooperatives, such as this sewing cooperative]]
	At the end of [[World War II]], the anarchist movement was severely weakened.{{sfn|Marshall|1993|p=xi}} However, the 1960s witnessed a revival of anarchism likely caused by a perceived failure of [[Marxism–Leninism]] and tensions built by the [[Cold War]].{{sfn|Marshall|1993|p=539}} During this time, anarchism took root in other movements critical towards both the state and capitalism, such as the [[Anti-nuclear movement|anti-nuclear]], [[Environmental movement|environmental]] and [[Peace movement|pacifist movements]], the [[New Left]], and the [[counterculture of the 1960s]].{{sfn|Marshall|1993|pp=xi, 539}} Anarchism became associated with [[punk subculture]], as exemplified by bands such as [[Crass]] and the [[Sex Pistols]],{{sfn|Marshall|1993|pp=493–494}} and the established [[feminist]] tendencies of [[anarcha-feminism]] returned with vigour during the [[second wave of feminism]].{{sfn|Marshall|1993|pp=556–557}}

	Around the turn of the 21st century, anarchism grew in popularity and influence within [[anti-war]], [[anti-capitalist]], and [[anti-globalisation movement]]s.{{sfn|Rupert|2006|p=66}} Anarchists became known for their involvement in protests against the [[World Trade Organization]], the [[Group of Eight]] and the [[World Economic Forum]]. During the protests, ''[[ad hoc]]'' leaderless anonymous cadres known as [[black bloc]]s engaged in [[riot]]ing, [[property destruction]], and violent confrontations with the [[police]]. Other organisational tactics pioneered in this time include [[security culture]], [[affinity group]]s, and the use of decentralised technologies such as the internet. A significant event of this period was the confrontations at the [[1999 Seattle WTO protests|WTO conference in Seattle in 1999]].{{sfn|Rupert|2006|p=66}} Anarchist ideas have been influential in the development of the [[Zapatista Army of National Liberation|Zapatistas]] in Mexico and the Democratic Federation of Northern Syria, more commonly known as [[Rojava]], a ''[[de facto]]'' [[Permanent autonomous zone|autonomous region]] in northern [[Syria]].{{sfn|Ramnath|2019|p=691}}

	== Thought ==
	{{main|Anarchist schools of thought}}
	[[Anarchist schools of thought]] have been generally grouped into two main historical traditions, [[social anarchism]] and [[individualist anarchism]], owing to their different origins, values and evolution.{{sfnm|1a1=McLean|1a2=McMillan|1y=2003|1loc=&quot;Anarchism&quot;|2a1=Ostergaard|2y=2003|2p=14|2loc=&quot;Anarchism&quot;}} The individualist current emphasises [[negative liberty]] in opposing restraints upon the free individual, while the social current emphasises [[positive liberty]] in aiming to achieve the free potential of society through equality and [[social ownership]].{{sfn|Harrison|Boyd|2003|p=251}} In a chronological sense, anarchism can be segmented by the classical currents of the late 19th century, and the post-classical currents (such as [[anarcha-feminism]], [[green anarchism]] and [[post-anarchism]]) developed thereafter.{{sfn|Levy|Adams|2019|p=9}}

	Beyond the specific factions of anarchist movements which constitute political anarchism lies [[philosophical anarchism]], which holds that the state lacks [[Morality|moral legitimacy]], without necessarily accepting the imperative of revolution to eliminate it.{{sfn|Egoumenides|2014|p=2}} A component especially of individualist anarchism,{{sfnm|1a1=Ostergaard|1y=2006|1p=12|2a1=Gabardi|2y=1986|2pp=300–302}} philosophical anarchism may tolerate the existence of a [[minimal state]], but argues that citizens have no [[moral obligation]] to obey government when it conflicts with individual autonomy.{{sfn|Klosko|2005|p=4}} Anarchism pays significant attention to moral arguments since [[ethics]] have a central role in anarchist philosophy.{{sfn|Franks|2019|p=549}}

	One reaction against [[sectarianism]] within the anarchist milieu was [[anarchism without adjectives]], a call for [[toleration]] and unity among anarchists first adopted by [[Fernando Tarrida del Mármol]] in 1889 in response to the bitter debates of anarchist theory at the time.{{sfn|Avrich|1996|p=6}} Despite separation, the various anarchist schools of thought are not seen as distinct entities, but as tendencies that intermingle.{{sfn|Marshall|1993|pp=1–6}}

	Anarchism is usually placed on the [[Far-left politics|far-left]] of the [[political spectrum]].{{sfnm|1a1=Brooks|1y=1994|1p=xi|2a1=Kahn|2y=2000|3a1=Moynihan|3y=2007}} Much of its [[Anarchist economics|economics]] and [[Anarchist law|legal philosophy]] reflect [[anti-authoritarian]], [[anti-statist]], and [[libertarian]] interpretations of the [[Political radicalism|radical]] [[left-wing]] and [[socialist]] politics{{sfn|Guerin|1970|p=12|ps=: &quot;[A]narchism is really a synonym for socialism. The anarchist is primarily a socialist whose aim is to abolish the exploitation of man by man. Anarchism is only one of the streams of socialist thought, that stream whose main components are concern for liberty and haste to abolish the State.&quot;}} of [[Collectivist anarchism|collectivism]], [[Anarcho-communism|communism]], [[Individualist anarchism|individualism]], [[Mutualism (economic theory)|mutualism]], and [[Anarcho-syndicalism|syndicalism]], among other [[libertarian socialist]] economic theories.{{sfn|Guerin|1970|p=35|loc=&quot;Critique of authoritarian socialism&quot;|ps=: &quot;The anarchists were unanimous in subjecting authoritarian socialism to a barrage of severe criticism. At the time when they made violent and satirical attacks these were not entirely well founded, for those to whom they were addressed were either primitive or &quot;vulgar&quot; communists, whose thought had not yet been fertilized by Marxist humanism, or else, in the case of Marx and Engels themselves, were not as set on authority and state control as the anarchists made out.&quot;}} As anarchism does not offer a fixed body of doctrine from a single particular worldview,{{sfn|Marshall|1993|pp=14–17}} many [[History of anarchism|anarchist types and traditions]] exist, and varieties of anarchy diverge widely.{{sfn|Sylvan|2007|p=262}}

	=== Classical ===
	[[File:Portrait of Pierre Joseph Proudhon 1865.jpg|thumb|upright|[[Pierre-Joseph Proudhon]], the primary proponent of [[Mutualism (economic theory)|anarcho-mutualism]], who influenced many future [[Individualist anarchism|individualist anarchist]] and [[Social anarchism|social anarchist]] thinkers{{sfn|Wilbur|2019|p=216-218}}]]

	Inceptive currents among classical anarchist currents were [[Mutualism (economic theory)|mutualism]] and [[Individualist anarchism|individualism]]. They were followed by the major currents of social anarchism ([[Collectivist anarchism|collectivist]], [[Anarcho-communism|communist]], and [[Anarcho-syndicalism|syndicalist]]). They differ on organizational and economic aspects of their ideal society.{{sfn|Levy|Adams|2019|p=2}}

	[[Mutualism (economic theory)|Mutualism]] is an 18th-century economic theory that was developed into anarchist theory by [[Pierre-Joseph Proudhon]]. Its aims include [[Reciprocity (cultural anthropology)|reciprocity]], [[Free association (Marxism and anarchism)|free association]], voluntary [[contract]], [[federation]], and [[Monetary reform|credit and currency reform]] that would be regulated by a bank of the people.{{sfn|Wilbur|2019|pp=213–218}} Mutualism has been retrospectively characterised as ideologically situated between individualist and collectivist forms of anarchism.{{sfnm|1a1=Avrich|1y=1996|1p=6|2a1=Miller|2y=1991|2p=11}} Proudhon first characterised his goal as a &quot;third form of society, the synthesis of communism and property&quot;.{{sfn|Pierson|2013|p=187|ps=. The quote is from Proudhon's ''[[What is Property?]]'' (1840).}}

	[[Collectivist anarchism]], also known as anarchist collectivism or anarcho-collectivism,{{sfn|Morris|1993|p=76}} is a [[revolutionary socialist]] form of anarchism commonly associated with [[Mikhail Bakunin]].{{sfn|Shannon|2019|p=101}} Collectivist anarchists advocate [[collective ownership]] of the means of production, theorised to be achieved through violent revolution,{{sfn|Avrich|1996|pp=3–4}} and that workers be paid according to time worked, rather than goods being distributed according to need as in communism. Collectivist anarchism arose alongside [[Marxism]], but rejected the [[dictatorship of the proletariat]] despite the stated Marxist goal of a collectivist [[stateless society]].{{sfnm|1a1=Heywood|1y=2017|1pp=146–147|2a1=Bakunin|2y=1990|2ps=: &quot;They [the Marxists] maintain that only a dictatorship—their dictatorship, of course—can create the will of the people, while our answer to this is: No dictatorship can have any other aim but that of self-perpetuation, and it can beget only slavery in the people tolerating it; freedom can be created only by freedom, that is, by a universal rebellion on the part of the people and free organization of the toiling masses from the bottom up.&quot;}} [[Anarcho-communism]], also known as anarchist-communism, communist anarchism, and libertarian communism, is a theory of anarchism that advocates a [[communist society]] with [[common ownership]] of the means of production,{{sfn|Mayne|1999|p=131}} [[direct democracy]], and a [[Horizontalidad|horizontal network]] of [[voluntary association]]s and [[workers' council]]s with production and consumption based on the guiding principle: &quot;[[From each according to his ability, to each according to his need]]&quot;.{{sfnm|1a1=Marshall|1y=1993|1p=327|2a1=Turcato|2y=2019|2pp=237–323}} Anarcho-communism developed from radical socialist currents after the [[French Revolution]],{{sfn|Graham|2005}} but it was first formulated as such in the Italian section of the [[First International]].{{sfn|Pernicone|2009|pp=111–113}} It was later expanded upon in the theoretical work of [[Peter Kropotkin]].{{sfn|Turcato|2019|p=239–244}}

	[[Anarcho-syndicalism]], also referred to as revolutionary syndicalism, is a branch of anarchism that views [[labour syndicate]]s as a potential force for revolutionary social change, replacing capitalism and the state with a new society democratically self-managed by workers. The basic principles of anarcho-syndicalism are workers' [[solidarity]], [[direct action]], and [[workers' self-management]].{{sfn|van der Walt|2019|p=249}}

	[[Individualist anarchism]] refers to several traditions of thought within the anarchist movement that emphasise the [[individual]] and their [[Will (philosophy)|will]] over any kinds of external determinants.{{sfn|Ryley|2019|p=225}}  Early influences on individualist forms of anarchism include [[William Godwin]], [[Max Stirner]] and [[Henry David Thoreau]]. Through many countries, individualist anarchism attracted a small yet diverse following of Bohemian artists and intellectuals{{sfn|Marshall|1993|p=440}} as well as young anarchist outlaws in what became known as [[illegalism]] and [[individual reclamation]].{{sfnm|1a1=Imrie|1y=1994|2a1=Parry|2y=1987|2p=15}}

	=== Post-classical and contemporary ===
	{{main|Contemporary anarchism}}
	[[File:Jarach and Zerzan.JPG|thumb|[[Lawrence Jarach]] (left) and [[John Zerzan]] (right), two prominent [[Contemporary anarchism|contemporary anarchist]] authors, with Zerzan being a prominent voice within [[anarcho-primitivism]] and Jarach a noted advocate of [[post-left anarchy]]]]

	Anarchist principles undergird contemporary radical social movements of the left. Interest in the anarchist movement developed alongside momentum in the [[anti-globalization movement]],{{sfn|Evren|2011|p=1}} whose leading activist networks were anarchist in orientation.{{sfn|Evren|2011|p=2}} As the movement shaped 21st century radicalism, wider embrace of anarchist principles signaled a revival of interest.{{sfn|Evren|2011|p=2}} Contemporary news coverage which emphasizes [[black bloc]] demonstrations has reinforced anarchism's historical association with chaos and violence, although its publicity has also led more scholars to engage with the anarchist movement.{{sfn|Evren|2011|p=1}} Anarchism has continued to generate many philosophies and movements—at times eclectic, drawing upon various sources, and [[Syncretic politics|syncretic]], combining disparate concepts to create new philosophical approaches.{{sfn|Perlin|1979}} The [[anti-capitalist]] tradition of classical anarchism has remained prominent within contemporary currents.{{sfn|Williams|2018|p=4}}

	Various anarchist groups, tendencies, and schools of thought exist today, making it difficult to describe contemporary anarchist movement.{{sfn|Franks|2013|pp=385–386}} While theorists and activists have established &quot;relatively stable constellations of anarchist principles&quot;, there is no consensus on which principles are core. As a result, commentators describe multiple &quot;anarchisms&quot; (rather than a singular &quot;anarchism&quot;) in which common principles are shared between schools of anarchism while each group prioritizes those principles differently. For example, gender equality can be a common principle but ranks as a higher priority to [[anarcha-feminists]] than [[anarchist communists]].{{sfn|Franks|2013|p=386}} Anarchists are generally committed against coercive authority in all forms, namely &quot;all centralized and hierarchical forms of government (e.g., monarchy, representative democracy, state socialism, etc.), economic class systems (e.g., capitalism, Bolshevism, feudalism, slavery, etc.), autocratic religions (e.g., fundamentalist Islam, Roman Catholicism, etc.), patriarchy, heterosexism, white supremacy, and imperialism&quot;.{{sfn|Jun|2009|pp=507–508}} However, anarchist schools disagree on the methods by which these forms should be opposed.{{sfn|Jun|2009|p=507}}

	== Tactics ==
	Anarchists' tactics take various forms but in general serve two major goals—first, to oppose [[the Establishment]]; and second, to promote anarchist ethics and reflect an anarchist vision of society, illustrating the unity of means and ends.{{sfn|Williams|2019|pp=107–108}} A broad categorization can be made between aims to destroy oppressive states and institutions by revolutionary means, and aims to change society through evolutionary means.{{sfn|Williams|2018|pp=4–5}} Evolutionary tactics [[Nonviolence|reject violence]] and take a gradual approach to anarchist aims, though there is significant overlap between the two.{{sfn|Kinna|2019|p=125}}

	Anarchist tactics have shifted during the course of the last century. Anarchists during the early 20th century focused more on strikes and militancy, while contemporary anarchists use a broader array of approaches.{{sfn|Williams|2019|p=112}}

	=== Classical era tactics ===
	[[File:McKinleyAssassination.jpg|thumb|The relationship between [[anarchism and violence]] is a controversial subject among anarchists. Pictured is [[Leon Czolgosz]] [[assassination of William McKinley|assassinating U.S. President William McKinley]].]]
	During the classical era, anarchists had a militant tendency. Not only did they confront state armed forces (as in [[Spain]] and [[Ukraine]]) but some of them also employed [[terrorism]] as [[propaganda of the deed]]. Assassination attempts were carried out against heads of state, some of which were successful. Anarchists also took part in [[revolution]]s.{{sfn|Williams|2019|pp=112–113}} Anarchist perspectives towards violence have always been perplexing and controversial.{{sfn|Carter|1978|p=320}} On one hand, [[Anarcho-pacifism|anarcho-pacifists]] point out the unity of means and ends.{{sfn|Fiala|2017}} On the other hand, other anarchist groups advocate direct action, a tactic which can include acts of [[sabotage]] or even acts of terrorism. This attitude was quite prominent a century ago; seeing the state as a [[tyrant]], some anarchists believed that they had every right to oppose its [[oppression]] by any means possible.{{sfn|Kinna|2019|pp=116–117}} [[Emma Goldman]] and [[Errico Malatesta]], who were proponents of limited use of violence, argued that violence is merely a reaction to state violence as a [[necessary evil]].{{sfn|Carter|1978|pp=320–325}}

	Anarchists took an active role in [[Strike action|strikes]], although they tended to be antipathetic to formal [[syndicalism]], seeing it as [[Reformism|reformist]]. They saw it as a part of the movement which sought to overthrow the [[State (polity)|state]] and [[capitalism]].{{sfn|Williams|2019|p=113}} Anarchists also reinforced their propaganda within the arts, some of whom practiced [[Naturism|nudism]]. They also built communities which were based on [[friendship]]. They were also involved in the [[News media|press]].{{sfn|Williams|2019|p=114}}

	=== Revolutionary tactics ===
	In the current era, Italian anarchist [[Alfredo Bonanno]], a proponent of [[insurrectionary anarchism]], has reinstated the debate on violence by rejecting the nonviolence tactic adopted since the late 19th century by Kropotkin and other prominent anarchists afterwards. Both Bonanno and the French group [[The Invisible Committee]] advocate for small, informal affiliation groups, where each member is responsible for their own actions but works together to bring down oppression utilizing sabotage and other violent means against state, capitalism and other enemies. Members of The Invisible Committee were arrested in 2008 on various charges, terrorism included.{{sfn|Kinna|2019|pp=134–135}}

	Overall, today's anarchists are much less violent and militant than their ideological ancestors. They mostly engage in confronting the police during demonstrations and riots, especially in countries like [[Anarchism in Canada|Canada]], [[Anarchism in Mexico|Mexico]] or [[Anarchism in Greece|Greece]]. Μilitant [[black bloc]] protest groups are known for clashing with the police.{{sfn|Williams|2019|p=115}} However, anarchists not only clash with state operators; they also engage in the struggle against fascists and racists, taking [[Anti-fascism|anti-fascist action]] and mobilizing to prevent hate rallies from happening.{{sfn|Williams|2019|p=117}}

	=== Evolutionary tactics ===
	Anarchists commonly employ [[direct action]]. This can take the form of disrupting and protesting against unjust [[hierarchy]], or the form of self-managing their lives through the creation of counter-institutions such as communes and non-hierarchical collectives.{{sfn|Williams|2018|pp=4–5}} Often, decision-making is handled in an anti-authoritarian way, with everyone having equal say in each decision, an approach known as [[Horizontalidad|horizontalism]].{{sfn|Williams|2019|pp=109–117}} Contemporary-era anarchists have been engaging with various [[grassroots]] movements that are not explicitly anarchist but are more or less based on horizontalism, respecting personal autonomy, and participating in mass activism such as strikes and demonstrations. The newly coined term &quot;small-a anarchism&quot;, in contrast with the &quot;big-A anarchism&quot; of the classical era, signals their tendency not to base their thoughts and actions on classical-era anarchism or to refer to Kropotkin or Proudhon to justify their opinions. They would rather base their thought and praxis on their own experience, which they will later theorize.{{sfn|Kinna|2019|pp=145–149}}

	The decision-making process of small [[Affinity group|affinity anarchist groups]] play a significant tactical role.{{sfn|Williams|2019|pp=109, 119}} Anarchists have employed various methods in order to build a rough consensus among members of their group, without the need of a leader or a leading group. One way is for an individual from the group to play the role of facilitator to help achieve a consensus without taking part in the discussion themselves or promoting a specific point. Minorities usually accept rough consensus, except when they feel the proposal contradicts anarchist goals, values, or ethics. Anarchists usually form small groups (5–20 individuals) to enhance autonomy and friendships among their members. These kind of groups more often than not interconnect with each other, forming larger networks. Anarchists still support and participate in strikes, especially [[Wildcat strike action|wildcat strikes]]; these are leaderless strikes not organised centrally by a syndicate.{{sfn|Williams|2019|p=119–121}}

	Anarchists have gone [[World Wide Web|online]] to spread their message. As in the past, newspapers and journals are used; however, because of distributional and other difficulties, anarchists have found it easier to create websites, hosting electronic libraries and other portals.{{sfn|Williams|2019|pp=118–119}} Anarchists were also involved in developing various software that are available for free. The way these hacktivists work to develop and distribute resembles the anarchist ideals, especially when it comes to preserving user's privacy from state surveillance.{{sfn|Williams|2019|pp=120–121}}

	Anarchists organize themselves to [[Squatting|squat]] and reclaim public spaces. During important events such as protests and when spaces are being occupied, they are often called [[Temporary Autonomous Zone]]s (TAZ), spaces where [[surrealism]], poetry and art are blended to display the anarchist ideal.{{sfnm|1a1=Kinna|1y=2019|1p=139|2a1=Mattern|2y=2019|2p=596|3a1=Williams|3y=2018|3pp=5–6}} As seen by anarchists, squatting is a way to regain urban space from the capitalist market, serving pragmatical needs, and is also seen an exemplary direct action.{{sfnm|1a1=Kinna|1y=2012|1p=250|2a1=Williams|2y=2019|2p=119}} Acquiring space enables anarchists to experiment with their ideas and build social bonds.{{sfn|Williams|2019|p=122}} Adding up these tactics, and having in mind that not all anarchists share the same attitudes towards them, along with various forms of protesting at highly symbolic events, make up a [[Carnivalesque|carnivalesque atmosphere]] that is part of contemporary anarchist vividity.{{sfn|Morland|2004|p=37–38}}

	== Key issues ==
	{{main|Issues in anarchism}}
	&lt;!-- In the interest of restricting article length, please limit this section to two or three short paragraphs and add any substantial information to the main Issues in anarchism article. Thank you. --&gt;
	As anarchism is a [[philosophy]] that embodies many diverse attitudes, tendencies, and schools of thought, and disagreement over questions of values, ideology, and tactics is common, its diversity has led to widely different uses of identical terms among different anarchist traditions, which has created a number of [[definitional concerns in anarchist theory]]. For instance, the compatibility of [[Anarchism and capitalism|capitalism]],{{sfnm|1a1=Marshall|1y=1993|1p=565|1ps=: &quot;In fact, few anarchists would accept the 'anarcho-capitalists' into the anarchist camp since they do not share a concern for economic equality and social justice, Their self-interested, calculating market men would be incapable of practising voluntary co-operation and mutual aid. Anarcho-capitalists, even if they do reject the State, might therefore best be called right-wing libertarians rather than anarchists.&quot;|2a1=Honderich|2y=1995|2p=31|3a1=Meltzer|3y=2000|3p=50|3ps=: &quot;The philosophy of &quot;anarcho-capitalism&quot; dreamed up by the &quot;libertarian&quot; [[New Right]], has nothing to do with Anarchism as known by the Anarchist movement proper.&quot;|4a1=Goodway|4y=2006|4p=4|4ps=: &quot;'Libertarian' and 'libertarianism' are frequently employed by anarchists as synonyms for 'anarchist' and 'anarchism', largely as an attempt to distance themselves from the negative connotations of 'anarchy' and its derivatives. The situation has been vastly complicated in recent decades with the rise of anarcho-capitalism, 'minimal statism' and an extreme right-wing laissez-faire philosophy advocated by such theorists as Murray Rothbard and Robert Nozick and their adoption of the words 'libertarian' and 'libertarianism'. It has therefore now become necessary to distinguish between their right libertarianism and the left libertarianism of the anarchist tradition.&quot;|5a1=Newman|5y=2010|5p=53|5ps=: &quot;It is important to distinguish between anarchism and certain strands of right-wing libertarianism which at times go by the same name (for example, Murray Rothbard's anarcho-capitalism).&quot;}} [[Anarchism and nationalism|nationalism]] and [[Anarchism and religion|religion]] with anarchism is widely disputed. Similarly, anarchism enjoys complex relationships with ideologies such as [[Anarchism and Marxism|Marxism]], [[Issues in anarchism#Communism|communism]], [[collectivism]] and [[trade unionism]]. Anarchists may be motivated by [[humanism]], [[God|divine authority]], [[enlightened self-interest]], [[Veganarchism|veganism]], or any number of alternative ethical doctrines. Phenomena such as [[civilisation]], [[technology]] (e.g. within [[anarcho-primitivism]]) and the [[Issues in anarchism#Participation in statist democracy|democratic process]] may be sharply criticised within some anarchist tendencies and simultaneously lauded in others.{{sfn|De George|2005|pp=31–32}}

	=== Gender, sexuality and free love ===
	{{main|Free love}}
	{{see also|Anarchism and issues related to love and sex|Queer anarchism}}
	[[File:Emilearmand01.jpg|left|thumb|upright|[[Individualist anarchism in France|French individualist anarchist]] [[Émile Armand]] propounded the virtues of [[free love]] in the [[Anarchism in France|Parisian anarchist milieu]] of the early 20th century]]

	Gender and sexuality carry along them dynamics of hierarchy; anarchism is obliged to address, analyse and oppose the suppression of one's autonomy because of the dynamics that gender roles traditionally impose.{{sfn|Nicholas|2019|p=603}}

	A historical current that arose and flourished during 1890 and 1920 within anarchism was [[free love]]; in contemporary anarchism, this current survives as a tendency to support [[polyamory]] and [[queer anarchism]].{{sfnm|1a1=Nicholas|1y=2019|1p=611|2a1=Jeppesen|2a2=Nazar|2y=2012|2pp=175–176}} Free love advocates were against marriage, which they saw as a way of men imposing authority over women, largely because marriage law greatly favoured the power of men. The notion of free love, though, was much broader; it included critique of the established order that limited women's sexual freedom and pleasure.{{sfn|Jeppesen|Nazar|2012|pp=175–176}} Such free love movements contributed to the establishment of communal houses, where large groups of travelers, anarchists, and other activists slept in beds together.{{sfn|Jeppesen|Nazar|2012|p=177}} Free love had roots both in Europe and the United States. Some anarchists, however, struggled with the jealousy that arose from free love.{{sfn|Jeppesen|Nazar|2012|pp=175–177}} Anarchist feminists were advocates of free love, against marriage, were pro-choice (utilizing a contemporary term) and had a likewise agenda. Anarchist and non-anarchist feminists differed on [[suffrage]], but were nonetheless supportive of one another.{{sfn|Kinna2019|pp=166–167}}

	During the second half of the 20th century, anarchism intermingled with the [[Second-wave feminism|second wave of feminism]], radicalizing some currents of the feminist movement (and being influenced as well). By the latest decades of the 20th century, anarchists and feminists were advocating for the rights and autonomy of women, gays, queers and other marginalized groups, with some feminist thinkers suggesting a fusion of the two currents.{{sfn|Nicholas|2019|pp=609–611}} With the [[Third-wave feminism|third wave of feminism]], sexual identity and compulsory heterosexuality became a subject of study for anarchists, which yielded a [[Post-structuralism|post-structuralist]] critique of sexual normality.{{sfn|Nicholas|2019|pp=610–611}} However, some anarchists distanced themselves from this line of thinking, suggesting that it leaned towards individualism and was, therefore, dropping the cause of social liberation.{{sfn|Nicholas|2019|pp=616–617}}

	=== Anarchism and education ===
	{{main|Anarchism and education}}
	{|class=&quot;wikitable&quot; style=&quot;border: none; background: none; float: right;&quot;
	|+ Anarchist vs. statist perspectives on education&lt;br&gt;&lt;small&gt;Ruth Kinna (2019){{sfn|Kinna|2019|p=97}}&lt;/small&gt;
	|-
	!scope=&quot;col&quot;|
	!scope=&quot;col&quot;|Anarchist education
	!scope=&quot;col&quot;|State education
	|-
	|Concept || Education as self-mastery || Education as service
	|-
	|Management || Community based || State run
	|-
	|Methods || Practice-based learning || Vocational training
	|-
	|Aims || Being a critical member of society || Being a productive member of society
	|}
	The interest of anarchists in education stretches back to the first emergence of classical anarchism. Anarchists consider 'proper' education, which sets the foundations of the future autonomy of the individual and the society, to be an act of [[Mutual aid (organization theory)|mutual aid]].{{sfnm|1a1=Kinna|1y=2019|1pp=83–85|2a2=Suissa|2y=2019|2pp=514–515, 520}} Anarchist writers such as [[William Godwin|Willian Godwin]] and [[Max Stirner]] attacked both state education and private education as another means by which the ruling class replicate their privileges.{{sfnm|1a1=Suissa|1y=2019|1pp=514, 521|2a1=Kinna|2y=2019|2pp=83–86|3a1=Marshall|3y=1993|3p=222|3ps=. Max Stirner recorded his thoughts in an essay on education titled &quot;[[The False Principle of Our Education]]&quot;. William Godwin's thoughts can be found in his ''[[Political Justice]]''.}}

	In 1901, Catalan anarchist and free thinker [[Francisco Ferrer]] established the [[Ferrer movement|Escuela Moderna]] in Barcelona as an opposition to the established education system, which was dictated largely by the Catholic Church.{{sfn|Suissa|2019|pp=511–512}} Ferrer's approach was secular, rejecting both state and church involvement in the educational process, and gave pupils large amounts of autonomy in planning their work and attendance. Ferrer aimed to educate the working class and explicitly sought to foster [[class consciousness]] among students. The school closed after constant harassment by the state and Ferrer was later arrested. His ideas, however, formed the inspiration for a series of [[Modern School (United States)|modern schools]] around the world.{{sfn|Suissa|2019|pp=511–514}} Christian anarchist [[Leo Tolstoy]] also established a similar school, with its founding principle, according to Tolstoy, being that &quot;for education to be effective it had to be free&quot;.{{sfn|Suissa|2019|pp=517–518|ps=. For more, see Tolstoy essay's ''Education and Culture''.}} In a similar token, A. S. Neill founding what became [[Summerhill School]] in 1921, also declaring being free from coercion.{{sfn|Suissa|2019|pp=518–519}} 

	Anarchist education is based largely on the idea that a child's right to develop freely, without manipulation, ought to be respected, and that rationality will lead children to morally good conclusions. However, there has been little consensus among anarchist figures as to what constitutes manipulation; Ferrer, for example, believed that moral indoctrination was necessary and explicitly taught pupils that equality, liberty, and social justice were not possible under capitalism (along with other critiques of nationalism and government).{{sfn|Suissa|2019|pp=519–522}}{{Sfn|Avrich|1980|p=|pp=3–33}} 

	Late 20th century and contemporary anarchist writers (such as [[Colin Ward]], [[Herbert Read]] and [[Paul Goodman]]) intensified and expanded the anarchist critique of state education, largely focusing on the need for a system that focuses on children's creativity rather than on their ability to attain a career or participate in [[Consumerism|consumer society]].{{sfn|Kinna|2019|pp=89–96}} Contemporary anarchists, such as [[Colin Ward]], have further argued that state education serves to perpetuate socio-economic inequality.{{Sfn|Ward|1973|p=|pp=39–48}}

	While few anarchist education institutions have survived to the modern day, major tenets of anarchist schools, such as respect for child autonomy and relying on reasoning rather than indoctrination as a teaching method, have spread among mainstream educational institutions.{{sfn|Suissa|2019|pp=523–526|ps=. Suissa names three schools as explicitly anarchists schools, namely the Free Skool Santa Cruz in the United States which is part of a wider American-Canadian network of schools; Self-Managed Learning College in Brighton, England; and Paideia School in Spain.}}

	=== Anarchism and the state ===
	&lt;!-- Important! Strive to explain how anarchists perceive authority and oppression and why they reject them. Jun (2019), p. 41. --&gt;
	Objection to the [[State (polity)|state]] and its institutions is a ''[[sine qua non]]'' of anarchism.{{sfnm|1a1=Carter|1y=1971|1p=14|2a1=Jun|2y=2019|2pp=29–30}} Anarchists consider the state as a tool of domination and believe it to be illegitimate regardless of its political tendencies. Instead of people being able to control the aspects of their life, major decisions are taken by a small elite. Authority ultimately rests solely on power, regardless of whether that power is [[Open government|open]] or [[Transparency (behavior)|transparent]], as it still has the ability to coerce people. Another anarchist argument against states is that the people constituting a government, even the most altruistic among officials, will unavoidably seek to gain more power, leading to corruption. Anarchists consider the idea that the state is the collective will of the people to be an unachievable fiction, due to the fact that the [[ruling class]] is distinct from the rest of society.{{sfn|Jun|2019|pp=32–38}}

	=== Anarchism and art ===
	[[File:Apple Harvest by Camille Pissarro.jpg|thumb|340px|''Les chataigniers a Osny'' by anarchist painter [[Camille Pissarro]]. 19th century [[Neo-Impressionism|neo-impressionist movement]] had an ecological aesthetic and offered an example of an anarchist perception of the road towards socialism.{{sfn|Antliff|1998|p=78}} In this specific painting, note how the blending of aestetic and social harmony is prefiguring an ideal anarchistic agrarian community.{{sfn|Antliff|1998|p=99}}]]{{See also|Anarchism and the arts}}{{Expand section|date=June 2020}}
	The connection between anarchism and art was quite profound during the classical era of anarchism, especially among artistic currents that were developing during that era, such as futurists, surrealists, and others,{{sfn|Mattern|2019|p=592}} while in literature anarchism was mostly associated with [[New Apocalypse]] and [[Neo Romanticism]] movements.{{snf|Gifford|2019|p=577}} In music anarchism has been associated with music scenes such as Punk.&lt;ref&gt;[http://www.allmusic.com/style/anarchist-punk-ma0000011967 Anarchist Punk | Music Highlights | AllMusic&lt;!-- Bot generated title --&gt;]&lt;/ref&gt; Anarchists such as Leo Tolstoy and Herbert Read argued that the border between the artist and the non-artist, what separates art from a daily act, is a construct produced by the alienation caused by capitalism, and it prevents humans from living a joyful life.{{sfn|Mattern|2019|pp=592–593}} 

	Other anarchists advocated for or used art as a means to achieve anarchist ends.{{sfn|Mattern|2019|p=593}} In his book Breaking the Spell: A History of Anarchist Filmmakers, Videotape Guerrillas, and Digital Ninjas Chris Robé claims that &quot;anarchist-inflected practices have increasingly structured movement-based video activism.&quot;{{sfn|Robé|2017|p=44}} 

	Three overlapping properties made art useful to anarchists: It could depict a critique of existing society and hierarchies; it could serve as a prefigurative tool to reflect the anarchist ideal society, and also it could turn into a means of direct action, in protests for example. As it appeals to both emotion and reason, art could appeal to the &quot;whole human&quot; and have a powerful effect.{{sfn|Mattern|2019|pp=593–596}}

	== Criticism{{anchor|Criticisms}} ==
	Philosophy lecturer Andrew G. Fiala has listed five main arguments against anarchism. Firstly, he notes that anarchism is related to violence and destruction, not only in the pragmatic world (i.e. at protests) but in the world of ethics as well. The second argument is that it is impossible for a society to function without a state or something like a state, acting to protect citizens from criminality. Fiala takes ''[[Leviathan (Hobbes book)|Leviathan]]'' from [[Thomas Hobbes]] and the [[night-watchman state]] from philosopher [[Robert Nozick]] as examples. Thirdly, anarchism is evaluated as unfeasible or utopian since the state can not be defeated practically; this line of arguments most often calls for political action within the system to reform it. The fourth argument is that anarchism is self-contradictory since while it advocates for no-one to ''archiei'', if accepted by the many, then anarchism will turn into the ruling political theory. In this line of criticism also comes the self contradiction that anarchist calls for collective action while anarchism endorses the autonomy of the individual and hence no collective action can be taken. Lastly, Fiala mentions a critique towards philosophical anarchism, of being ineffective (all talk and thoughts) and in the meantime capitalism and bourgeois class remains strong.{{sfn|Fiala|2017|loc=&quot;4. Objections and Replies&quot;}}

	Philosophical anarchism has met the criticism of members of academia, following the release of pro-anarchist books such as [[A. John Simmons]]' ''Moral Principles and Political Obligations'' (1979).{{sfn|Klosko|1999|p=536}} Law professor William A. Edmundson authored an essay arguing against three major philosophical anarchist principles, which he finds fallacious; Edmundson claims that while the individual does not owe a normal state{{ambiguous|date=June 2020}} a duty of obedience, this does not imply that anarchism is the inevitable conclusion, and the state is still morally legitimate.{{sfnm|1a1=Klosko|1y=1999|1p=536|2a1=Kristjánsson|2y=2000|2p=896}}

	== See also ==
	{{Portal|Anarchism|Libertarianism}}
	* [[:Category:Anarchism by country|Anarchism by country]]
	* [[Governance without government]]
	* [[List of political ideologies#Anarchism|List of anarchist political ideologies]]
	* [[List of books about anarchism]]

	== References ==
	=== Citations ===
	{{reflist|25em}}

	=== Sources ===
	; Primary sources
	{{refbegin|35em|indent=yes}}
	* {{cite book |last=Bakunin|first=Mikhail|author-link=Mikhail Bakunin|title=Statism and Anarchy|title-link=Statism and Anarchy|year=1990|orig-year=1873|publisher=Cambridge University Press|location=Cambridge, England |series=Cambridge Texts in the History of Political Thought|translator-last=Shatz|translator-first=Marshall|isbn=978-0-521-36182-8|oclc=20826465|lccn=89077393|doi=10.1017/CBO9781139168083|ref=harv|editor1-last=Shatz|editor1-first=Marshall}}
	{{refend}}

	; Secondary sources
	{{refbegin|35em|indent=yes}}
	* {{cite journal|last=Adams|first=Matthew|date=14 January 2014|title=The Possibilities of Anarchist History: Rethinking the Canon and Writing History|url=https://journals.uvic.ca/index.php/adcs/article/view/17138|journal=Anarchist Developments in Cultural Studies|volume=2013.1: Blasting the Canon|pages=33–63|accessdate=17 December 2019|via=University of Victoria Libraries|ref=harv}}
	* {{cite journal|last=Antliff|first=Mark|year=1998|title=Cubism, Futurism, Anarchism: The 'Aestheticism' of the &quot;Action d'art&quot; Group, 1906–1920|url=|journal=Oxford Art Journal|volume=21|issue=2|pages=101–120|ref=harv|jstor=1360616|doi=10.1093/oxartj/21.2.99}}
	* {{cite journal|last=Anderson|first=Benedict|author-link=Benedict Anderson|title=In the World-Shadow of Bismarck and Nobel|journal=[[New Left Review]]|volume=2|issue=28|pages=85–129|year=2004|url=http://newleftreview.org/II/28/benedict-anderson-in-the-world-shadow-of-bismarck-and-nobel|access-date=7 January 2016|archive-url=https://web.archive.org/web/20151219130121/http://newleftreview.org/II/28/benedict-anderson-in-the-world-shadow-of-bismarck-and-nobel|archive-date=19 December 2015|ref=harv|via=}}
	* {{cite book|last=Avrich|first=Paul|title=Anarchist Voices: An Oral History of Anarchism in America|year=1996|publisher=Princeton University Press|isbn=978-0-691-04494-1|location=|pages=|ref=harv}}
	* {{cite book|last=Avrich|first=Paul|author-link=Paul Avrich|title=The Russian Anarchists|publisher=[[AK Press]]|location=Stirling|pages=|year=2006|isbn=978-1-904859-48-2|ref=harv|title-link=The Russian Anarchists}}
	* {{Cite book|last=Avrich|first=Paul|url=https://books.google.com/books?id=X6X_AwAAQBAJ&amp;pg=PA3#v=onepage|title=The Modern School Movement: Anarchism and Education in the United States|publisher=Princeton University Press|year=1980|isbn=978-1-4008-5318-2|location=|pages=3–33|oclc=489692159}}
	* {{cite book|last=Bantman|first=Constance|chapter=The Era of Propaganda by the Deed|location=|pages=371–388|editor1-last=Levy|editor1-first=Carl|editor1-link=Carl Levy (political scientist)|editor2-last=Adams|editor2-first=Matthew S.|title=The Palgrave Handbook of Anarchism|year=2019|publisher=[[Springer Publishing]]|isbn=978-3-319-75620-2|ref=harv}}
	* {{cite book|ref=harv|last=Bates|first=David|editor=Paul Wetherly|location=|title=Political Ideologies|chapter-url=https://books.google.com/books?id=uXfJDgAAQBAJ&amp;pg=PA128|year=2017|publisher=Oxford University Press|isbn=978-0-19-872785-9|pages=|chapter=Anarchism}}
	* {{cite book|last=Bolloten|first=Burnett|author-link=Burnett Bolloten|title=The Spanish Civil War: Revolution and Counterrevolution|publisher=University of North Carolina Press|year=1984|isbn=978-0-8078-1906-7|location=|pages=|ref=harv}}
	* {{cite book|last=Brooks|first=Frank H.|year=1994|title=The Individualist Anarchists: An Anthology of Liberty (1881–1908)|publisher=Transaction Publishers|isbn=978-1-56000-132-4|location=|pages=|ref=harv}}
	* {{cite book|last=Carter|first=April|author-link=April Carter|title=The Political Theory of Anarchism|url=https://books.google.com/books?id=3mlWPgAACAAJ|year=1971|publisher=Routledge|isbn=978-0-415-55593-7|location=|pages=|ref=harv}}
	* {{cite journal|last=Carter|first=April|title=Anarchism and violence|url=|journal=Nomos|volume=19|pages=320–340|year=1978|publisher=American Society for Political and Legal Philosophy|ref=harv|via=|jstor=24219053}}
	* {{cite book|editor1-last=Chaliand|editor1-first=Gerard|editor2-last=Blin|editor2-first=Arnaud|title=The History of Terrorism: From Antiquity to Al-Quaeda|publisher=University of California Press|location=Berkeley, CA; Los Angeles, CA; London, England|pages=|year=2007|isbn=978-0-520-24709-3|oclc=634891265|ref=harv|url-access=registration|last=|first=|url=https://archive.org/details/historyofterrori00grar}}
	* {{cite book|last=Dirlik|first=Arif|title=Anarchism in the Chinese Revolution|publisher=University of California Press|location=Berkeley, CA|pages=|year=1991|isbn=978-0-520-07297-8|ref=harv}}
	* {{cite book|last=Dodson|first=Edward|title=The Discovery of First Principles|volume=2|location=|pages=|publisher=Authorhouse|year=2002|isbn=978-0-595-24912-1|ref=harv}}
	* {{cite book|last=Egoumenides|first=Magda|title=Philosophical Anarchism and Political Obligation|url=https://books.google.com/books?id=DMxgBwAAQBAJ|date=28 August 2014|publisher=Bloomsbury Academic|isbn=978-1-4411-4411-9|location=|pages=|ref=harv}}
	* {{cite book|last=Evren|first=Süreyyya|author-link=Süreyyya Evren|chapter=How New Anarchism Changed the World (of Opposition) after Seattle and Gave Birth to Post-Anarchism|location=|pages=1–19|editor-last1=Rousselle|editor-first1=Duane|editor-last2=Evren|editor-first2=Süreyyya|title=Post-Anarchism: A Reader|year=2011|isbn=978-0-7453-3086-0|publisher=[[Pluto Press]]|ref=harv}}
	* {{cite book|last=Fernández|first=Frank|title=Cuban Anarchism: The History of A Movement|year=2009|isbn=|location=|pages=|orig-year=2001|publisher=Sharp Press|ref=harv}}
	* {{Cite journal|last1=Franks|first1=Benjamin|authorlink=Benjamin Franks|editor-last1=Freeden|editor-first1=Michael|editor-last2=Stears|editor-first2=Marc|title=Anarchism|url=|journal=The Oxford Handbook of Political Ideologies|volume=|pages=385–404|date=August 2013|publisher=Oxford University Press|language=en|doi=10.1093/oxfordhb/9780199585977.013.0001|ref=harv}}
	* {{cite book|last1=Franks|first1=Benjamin|authorlink=Benjamin Franks|chapter=Anarchism and Ethics|location=|pages=549–570|editor1-last=Levy|editor1-first=Carl|editor1-link=Carl Levy (political scientist)|editor2-last=Adams|editor2-first=Matthew S.|title=The Palgrave Handbook of Anarchism|year=2019|publisher=[[Springer Publishing]]|isbn=978-3-319-75620-2|ref=harv}}
	* {{cite book|last1=Jeppesen|first1=Sandra|last2=Nazar|first2=Holly|pages=|chapter=Genders and Sexualities in Anarchist Movements|editor=[[Ruth Kinna]]|location=|title=The Bloomsbury Companion to Anarchism|url=https://books.google.com/books?id=dNuoAwAAQBAJ|date=28 June 2012|publisher=Bloomsbury Publishing|isbn=978-1-4411-4270-2|ref=harv}}
	* {{cite journal|last1=Jun|first1=Nathan|title=Anarchist Philosophy and Working Class Struggle: A Brief History and Commentary|url=|journal=[[WorkingUSA]]|volume=12|issue=3|pages=505–519|date=September 2009|language=en|doi=10.1111/j.1743-4580.2009.00251.x|issn=1089-7011|ref=harv}}&lt;!-- to revisit, p. 508+ --&gt;
	* {{cite book|last=Jun|first=Nathan|chapter=The State|location=|pages=27–47|editor1-last=Levy|editor1-first=Carl|editor1-link=Carl Levy (political scientist)|editor2-last=Adams|editor2-first=Matthew S.|title=The Palgrave Handbook of Anarchism|year=2019|publisher=[[Springer Publishing]]|isbn=978-3-319-75620-2|ref=harv}}
	* {{cite journal|last=Gabardi|first=Wayne|year=1986|volume=80|issue=1|pages=300–302|doi=10.2307/1957102|jstor=446800|ref=harv|title=Anarchism. By David Miller. (London: J. M. Dent and Sons, 1984. Pp. 216. £10.95.)|url=|journal=American Political Science Review}}
	* {{cite book|last=Gifford|first=James|pages=|chapter=Literature and Anarchism|editor1-last=Levy|editor1-first=Carl|editor1-link=Carl Levy (political scientist)|location=|editor2-last=Adams|editor2-first=Matthew S.|title=The Palgrave Handbook of Anarchism|year=2019|publisher=[[Springer Publishing]]|isbn=978-3-319-75620-2|ref=harv}}
	* {{cite book|last=Goodway|first=David|author-link=David Goodway|title=Anarchist Seeds Beneath the Snow|publisher=Liverpool Press|year=2006|isbn=978-1-84631-025-6|location=|pages=|ref=harv}}
	* {{cite book|last=Graham|first=Robert|title=Anarchism: a Documentary History of Libertarian Ideas: from Anarchy to Anarchism|publisher=Black Rose Books|location=Montréal|pages=|year=2005|isbn=978-1-55164-250-5|author-link=Robert Graham (historian)|url=http://robertgraham.wordpress.com/anarchism-a-documentary-history-of-libertarian-ideas-volume-one-from-anarchy-to-anarchism-300ce-1939/|access-date=5 March 2011|archive-url=https://web.archive.org/web/20101130131904/http://robertgraham.wordpress.com/anarchism-a-documentary-history-of-libertarian-ideas-volume-one-from-anarchy-to-anarchism-300ce-1939/|archive-date=30 November 2010|ref=harv}}
	* {{cite book|last=Graham|first=Robert|author-link=Robert Graham (historian)|chapter=Anarchism and the First International|location=|pages=325–342|editor1-last=Levy|editor1-first=Carl|editor1-link=Carl Levy (political scientist)|editor2-last=Adams|editor2-first=Matthew S.|title=The Palgrave Handbook of Anarchism|chapter-url=https://books.google.com/books?id=SRyQswEACAAJ|year=2019|publisher=Springer|isbn=978-3-319-75620-2|ref=harv}}
	* {{cite book|last=Guerin|first=Daniel|author-link=Daniel Guérin|title=Anarchism: From Theory to Practice|url=http://theanarchistlibrary.org/library/daniel-guerin-anarchism-from-theory-to-practice|year=1970|isbn=|location=|pages=|publisher=Monthly Review Press|ref=harv}}
	* {{cite book|ref=harv|last1=Harrison|first1=Kevin|last2=Boyd|first2=Tony|title=Understanding Political Ideas and Movements|url=https://books.google.com/books?id=5qrJCgAAQBAJ|date=5 December 2003|publisher=Manchester University Press|isbn=978-0-7190-6151-6|location=|pages=}}
	* {{cite book|last=Heywood|first=Andrew|author-link=Andrew Heywood|title=Political Ideologies: An Introduction|url=https://books.google.com/books?id=Sy8hDgAAQBAJ&amp;pg=PA146|edition=6th|location=|pages=|date=16 February 2017|publisher=Macmillan International Higher Education|isbn=978-1-137-60604-4|ref=harv}}
	* {{cite book|last=Honderich|first=Ted|title=The Oxford Companion to Philosophy|url=https://archive.org/details/oxfordcompaniont00hond|url-access=registration|year=1995|publisher=Oxford University Press|isbn=978-0-19-866132-0|location=|pages=|ref=harv}}
	* {{cite article|last=Imrie|first=Doug|title=The Illegalists|year=1994|work=Anarchy: A Journal of Desire Armed|url=http://recollectionbooks.com/siml/library/illegalistsDougImrie.htm|url-status=dead|archive-url=https://web.archive.org/web/20150908072801/http://recollectionbooks.com/siml/library/illegalistsDougImrie.htm|archive-date=8 September 2015|access-date=9 December 2010|ref=harv}}
	* {{cite book|title=The Anarchists|last=Joll|first=James|author-link=James Joll|year=1964|publisher=Harvard University Press|isbn=978-0-674-03642-0|location=|pages=|ref=harv}}
	* {{cite journal|last=Kahn|first=Joseph|title=Anarchism, the Creed That Won't Stay Dead; The Spread of World Capitalism Resurrects a Long-Dormant Movement|url=|year=2000|journal=[[The New York Times]]|volume=|issue=5 August|pages=|ref=harv|via=}}
	* {{cite book|last=Kinna|first=Ruth|author-link=Ruth Kinna|title=The Bloomsbury Companion to Anarchism|year=2012|publisher=Bloomsbury Academic|isbn=978-1-62892-430-5|location=|pages=|ref=harv}}
	* {{cite book|last=Kinna|first=Ruth|author-link=Ruth Kinna|title=The Government of No One, The Theory and Practice of Anarchism|publisher=[[Penguin Random House]]|url=https://books.google.com/books?id=xzeGDwAAQBAJ|year=2019|isbn=978-0-241-39655-1|location=|pages=|ref=harv}}
	* {{cite journal|last=Klosko|first=George|title=More than Obligation - William A. Edmundson: Three Anarchical Fallacies: An Essay on Political Authority.- The Review of Politics|url=|journal=The Review of Politics|volume=61|issue=3|year=1999|issn=1748-6858|doi=10.1017/S0034670500028989|pages=536–538|ref=harv}}
	* {{cite book|last=Klosko|first=George|title=Political Obligations|url=https://books.google.com/books?id=ToHmfIj8d_gC|year=2005|publisher=Oxford University Press|isbn=978-0-19-955104-0|location=|pages=|ref=harv}}
	* {{cite journal|last=Kristjánsson|first=Kristján|title=Three Anarchical Fallacies: An Essay on Political Authority by William A. Edmundson|url=|journal=Mind|volume=109|issue=436|pp=896–900|year=2000|ref=harv|via=|jstor=2660038}}
	* {{cite book|last=Laursen|first=Ole Birk|chapter=Anti-Imperialism|location=|pages=149–168|editor1-last=Levy|editor1-first=Carl|editor1-link=Carl Levy (political scientist)|editor2-last=Adams|editor2-first=Matthew S.|title=The Palgrave Handbook of Anarchism|chapter-url=https://books.google.com/books?id=SRyQswEACAAJ|year=2019|publisher=Springer|isbn=978-3-319-75620-2|ref=harv}}
	* {{cite journal|last=Levy|first=Carl|s2cid=144317650|author-link=Carl Levy (political scientist)|title=Social Histories of Anarchism|journal=Journal for the Study of Radicalism|volume=4|issue=2|date=8 May 2011|issn=1930-1197|doi=10.1353/jsr.2010.0003|pages=1–44|ref=harv}}
	* {{cite book|last1=Levy|first1=Carl|last2=Adams|first2=Matthew S.|chapter=Introduction|location=|pages=1–23|editor1-last=Levy|editor1-first=Carl|editor1-link=Carl Levy (political scientist)|editor2-last=Adams|editor2-first=Matthew S.|title=The Palgrave Handbook of Anarchism|year=2019|publisher=[[Springer Publishing]]|isbn=978-3-319-75620-2|ref=harv}}
	* {{cite book|last=Long|first=Roderick T.|author-link=Roderick T. Long|editor-last1=Gaud|editor-first1=Gerald F.|location=|pages=|editor-last2=D'Agostino|editor-first2=Fred|title=The Routledge Companion to Social and Political Philosophy|url=https://books.google.com/books?id=z7MzEHaJgKAC|year=2013|publisher=Routledge|isbn=978-0-415-87456-4|ref=harv}}
	* {{cite book|last=Marshall|first=Peter|author-link=Peter Marshall (author)|title=Demanding the Impossible: A History of Anarchism|year=1993|publisher=PM Press|place=Oakland, CA|pages=|isbn=978-1-60486-064-1|ref=harv}}
	* {{cite book|last=Mattern|first=Mark|chapter=Anarchism and Art|location=|pages=589–602|editor1-last=Levy|editor1-first=Carl|editor1-link=Carl Levy (political scientist)|editor2-last=Adams|editor2-first=Matthew S.|title=The Palgrave Handbook of Anarchism|year=2019|publisher=[[Springer Publishing]]|isbn=978-3-319-75620-2|ref=harv}}
	* {{cite book|last=Mayne|first=Alan James|url=https://books.google.com/books?id=6MkTz6Rq7wUC&amp;pg=PA131&amp;dq=Communist+anarchism+believes+in+collective+ownership|title=From Politics Past to Politics Future: An Integrated Analysis of Current and Emergent Paradigms|year=1999|publisher=Greenwood Publishing Group|accessdate=20 September 2010|isbn=978-0-275-96151-0|location=|pages=|ref=harv}}
	* {{cite book|last=McLaughlin|first=Paul|title=Anarchism and Authority: A Philosophical Introduction to Classical Anarchism|url=https://we.riseup.net/assets/394498/paul-mclaughlin-anarchism-and-authority-a-philosophical-introduction-to-classical-anarchism-1.pdf|archive-url=https://web.archive.org/web/20180804180522/https://we.riseup.net/assets/394498/paul-mclaughlin-anarchism-and-authority-a-philosophical-introduction-to-classical-anarchism-1.pdf|archive-date=4 August 2018|publisher=[[Ashgate Publishing|Ashgate]]|location=Aldershot|pages=|date=28 November 2007|isbn=978-0-7546-6196-2|ref=harv}}
	* {{cite book|last1=Morland|first1=Dave|chapter=Anti-capitalism and poststructuralist anarchism|editor1=Jonathan Purkis|editor2=James Bowen|location=|pp=23–38|title=Changing Anarchism: Anarchist Theory and Practice in a Global Age|url=https://books.google.com/books?id=etb2UFzCBv4C|year=2004|publisher=Manchester University Press|isbn=978-0-7190-6694-8|ref=harv}}
	* {{cite book|last=Meltzer|first=Albert|author-link=Albert Meltzer|title=Anarchism: Arguments For and Against|url=https://archive.org/details/anarchism00albe|url-access=registration|date=1 January 2000|publisher=AK Press|isbn=978-1-873176-57-3|location=|pages=|ref=harv}}
	* {{cite book|last=Morris|first=Brian|title=Bakunin: The Philosophy of Freedom|url=https://books.google.com/books?id=GJXy5eCpPawC|date=January 1993|publisher=Black Rose Books|isbn=978-1-895431-66-7|location=|pages=|ref=harv}}
	* {{cite book|ref=harv|last=Morris|first=Christopher W.|title=An Essay on the Modern State|url=https://books.google.com/books?id=uuyJ9Bw8w7QC|year=2002|publisher=Cambridge University Press|isbn=978-0-521-52407-0|location=|pages=}}
	* {{cite journal|last=Moynihan|first=Colin|title=Book Fair Unites Anarchists. In Spirit, Anyway|url=|year=2007|journal=The New York Times|volume=|issue=16 April|pages=|ref=harv|via=}}
	* {{cite book|ref=harv|last=Moya|first=Jose C|editor=Geoffroy de Laforcade|location=|others=Kirwin R. Shaffer|title=In Defiance of Boundaries: Anarchism in Latin American History|chapter-url=https://books.google.com/books?id=ikt6AQAACAAJ|year=2015|publisher=[[University Press of Florida]]|isbn=978-0-8130-5138-3|pages=|chapter=Transference, culture, and critique The Circulation of Anarchist Ideas and Practices}}
	* {{cite book|last=Nettlau|first=Max|author-link=Max Nettlau|title=A Short History of Anarchism|year=1996|publisher=Freedom Press|isbn=978-0-900384-89-9|location=|pages=|ref=harv}}
	* {{cite book|last=Newman|first=Saul|title=The Politics of Postanarchism|url=https://books.google.com/books?id=SiqBiViUsOkC&amp;pg=PA43|year=2010|publisher=Edinburgh University Press|isbn=978-0-7486-3495-8|location=|pages=|ref=harv}}
	* {{cite book|last=Nicholas|first=Lucy|editor1-last=Levy|editor1-first=Carl|editor1-link=Carl Levy (political scientist)|location=|editor2-last=Adams|editor2-first=Matthew S.|title=The Palgrave Handbook of Anarchism|year=2019|publisher=Springer|isbn=978-3-319-75620-2|pages=|chapter=Gender and Sexuality|ref=harv}}
	* {{cite book|last=Nomad|first=Max|contribution=The Anarchist Tradition|editor1-last=Drachkovitch|editor1-first=Milorad M.|title=Revolutionary Internationals 1864–1943|publisher=Stanford University Press|location=|page=88|year=1966|isbn=978-0-8047-0293-5|ref=harv}}
	* {{cite book|last=Ostergaard|first=Geoffrey|author-link=Geoffrey Ostergaard|editor=William Outhwaite|location=|pages=|title=The Blackwell Dictionary of Modern Social Thought|url=https://books.google.com/books?id=JJmdpqJwkwwC&amp;pg=PA14|year=2006|publisher=Blackwell Publishing Ltd.|isbn=978-0-470-99901-1|ref=harv}}
	* {{cite book|last=Parry|first=Richard|title=The Bonnot Gang|url=https://archive.org/details/bonnotgang0000parr|url-access=registration|year=1987|publisher=Rebel Press|isbn=978-0-946061-04-4|location=|pages=|ref=harv}}
	* {{cite book|last=Perlin|first=Terry M.|year=1979|title=Contemporary Anarchism|url=https://books.google.com/books?id=mppLKlwHx7oC|publisher=Transaction Publishers|isbn=978-1-4128-2033-2|location=|pages=|ref=harv}}
	* {{cite book|last=Pernicone|first=Nunzio|title=Italian Anarchism, 1864–1892|url=https://books.google.com/books?id=3ttgjwEACAAJ|year=2009|publisher=Princeton University Press|isbn=978-0-691-63268-1|location=|pages=|ref=harv}}
	* {{cite book|ref=harv|last=Pierson|first=Christopher|title=Just Property: Enlightenment, Revolution, and History|url=https://books.google.com/books?id=7jvKDAAAQBAJ&amp;pg=PA187|year=2013|publisher=Oxford University Press|isbn=978-0-19-967329-2|location=|pages=}}
	* {{cite book|last=Ramnath|first=Maia|editor1-last=Levy|editor1-first=Carl|editor1-link=Carl Levy (political scientist)|location=|editor2-last=Adams|editor2-first=Matthew S.|title=The Palgrave Handbook of Anarchism|chapter-url=https://books.google.com/books?id=SRyQswEACAAJ|year=2019|publisher=Springer|isbn=978-3-319-75620-2|pages=|chapter=Non-Western Anarchisms and Postcolonialism|ref=harv}}
	* {{cite book|last=Robé|first=Chris|title=Breaking the Spell: A History of Anarchist Filmmakers, Videotape Guerrillas, and Digital Ninjas |url=https://www.researchgate.net/profile/Chris_Robe/publication/336710855_Breaking_the_Spell_A_History_of_Anarchist_Filmmakers_Videotape_Guerrillas_and_Digital_Ninjas/links/5dae619d92851c577eb971ce/Breaking-the-Spell-A-History-of-Anarchist-Filmmakers-Videotape-Guerrillas-and-Digital-Ninjas.pdf|year=2017|publisher=PM Press|isbn=978-1-629-63233-9}}
	* {{cite book|last=Rupert|first=Mark|title=Globalization and International Political Economy|publisher=Rowman &amp; Littlefield Publishers|location=Lanham|pages=|year=2006|isbn=978-0-7425-2943-4|ref=harv|url=https://archive.org/details/globalizationint00rupe}}
	* {{cite book|last=Ryley|first=Peter|chapter=Individualism|location=|pages=225–236|editor1-last=Levy|editor1-first=Carl|editor1-link=Carl Levy (political scientist)|editor2-last=Adams|editor2-first=Matthew S.|title=The Palgrave Handbook of Anarchism|year=2019|publisher=[[Springer Publishing]]|isbn=978-3-319-75620-2|ref=harv}}
	* {{cite book|last=Shannon|first=Deric|chapter=Anti-Capitalism and Libertarian Political Economy|location=|pages=91–106|editor1-last=Levy|editor1-first=Carl|editor1-link=Carl Levy (political scientist)|editor2-last=Adams|editor2-first=Matthew S.|title=The Palgrave Handbook of Anarchism|year=2019|publisher=[[Springer Publishing]]|isbn=978-3-319-75620-2|ref=harv}}
	* {{cite book|last=Skirda|first=Alexandre|title=Facing the Enemy: A History of Anarchist Organization From Proudhon to May 1968|publisher=AK Press|year=2002|isbn=978-1-902593-19-7|location=|pages=|title-link=Facing the Enemy|ref=harv}}
	* {{cite book|last=Sylvan|year=2007|first=Richard|section=Anarchism|editor=Robert E. Goodin|editor2=Philip Pettit|editor3=Thomas Pogge|title=A Companion to Contemporary Political Philosophy|edition=2nd|url=http://eltalondeaquiles.pucp.edu.pe/wp-content/uploads/2016/04/Robert-E--Goodin-Philip-Pettit-Thomas-W--Pogge-A-Companion-to-Contemporary-Political-Philosophy-2-Volume-Set-Blackwell-Companions-to-Philosophy-2007.pdf|archive-url=https://web.archive.org/web/20170517032711/http://eltalondeaquiles.pucp.edu.pe/wp-content/uploads/2016/04/Robert-E--Goodin-Philip-Pettit-Thomas-W--Pogge-A-Companion-to-Contemporary-Political-Philosophy-2-Volume-Set-Blackwell-Companions-to-Philosophy-2007.pdf|archive-date=17 May 2017|series=Blackwell Companions to Philosophy|volume=5|location=|pages=|publisher=Blackwell Publishing|isbn=978-1-4051-3653-2|author-link=Richard Sylvan|editor1-link=Robert E. Goodin|editor2-link=Philip Pettit|editor3-link=Thomas Pogge|ref=harv}}
	* {{cite book|last=Suissa|first=Judith|pages=|chapter=Anarchist Education|editor1-last=Levy|editor1-first=Carl|editor1-link=Carl Levy (political scientist)|location=|editor2-last=Adams|editor2-first=Matthew S.|title=The Palgrave Handbook of Anarchism|year=2019|chapter-url=https://books.google.com/books?id=SRyQswEACAAJ|publisher=[[Springer Publishing]]|isbn=978-3-319-75620-2|ref=harv}}
	* {{cite book|last=Thomas|first=Paul|title=Karl Marx and the Anarchists|publisher=Routledge &amp; Kegan Paul|location=London|pages=|year=1985|isbn=978-0-7102-0685-5|ref=harv}}
	* {{cite book|last=Turcato|first=Davide|pages=|chapter=Anarchist Communism|editor1-last=Levy|editor1-first=Carl|editor1-link=Carl Levy (political scientist)|location=|editor2-last=Adams|editor2-first=Matthew S.|title=The Palgrave Handbook of Anarchism|year=2019|publisher=[[Springer Publishing]]|isbn=978-3-319-75620-2|ref=harv}}
	* {{cite book|last=Van der Walt|first=Lucien|chapter=Syndicalism|location=|pages=249–264|editor1-last=Levy|editor1-first=Carl|editor1-link=Carl Levy (political scientist)|editor2-last=Adams|editor2-first=Matthew S.|title=The Palgrave Handbook of Anarchism|year=2019|publisher=[[Springer Publishing]]|isbn=978-3-319-75620-2|ref=harv}}
	* {{cite book|last=Ward|first=Colin|author-link=Colin Ward|title=Anarchism: A Very Short Introduction|url=https://books.google.com/books?id=nkgSDAAAQBAJ|date=21 October 2004|publisher=OUP Oxford|isbn=978-0-19-280477-8|location=|pages=|ref=harv}}
	* {{cite journal|last=Ward|first=Colin|date=1973|title=The Role of the State|url=|journal=Education Without Schools|volume=|pages=39–48|via=}}
	* {{cite book|last=Wilbur|first=Shawn|pages=|chapter=Mutualism|editor1-last=Levy|editor1-first=Carl|editor1-link=Carl Levy (political scientist)|location=|editor2-last=Adams|editor2-first=Matthew S.|title=The Palgrave Handbook of Anarchism|year=2019|publisher=[[Springer Publishing]]|isbn=978-3-319-75620-2|ref=harv}}
	* {{cite journal|last=Williams|first=Dana M.|title=Contemporary anarchist and anarchistic movements|url=|journal=Sociology Compass|publisher=Wiley|volume=12|issue=6|pages=e12582|year=2018|issn=1751-9020|doi=10.1111/soc4.12582|ref=harv}}
	* {{cite book|last=Williams|first=Dana M.|pages=|chapter=Tactics: Conceptions of Social Change, Revolution, and Anarchist Organisation|editor1-last=Levy|editor1-first=Carl|editor1-link=Carl Levy (political scientist)|location=|editor2-last=Adams|editor2-first=Matthew S.|title=The Palgrave Handbook of Anarchism|year=2019|publisher=[[Springer Publishing]]|isbn=978-3-319-75620-2|ref=harv}}
	* {{refend}}

	; Tertiary sources
	{{refbegin|35em|indent=yes}}
	* {{cite web|last=Coutinho|first=Steve|title=Zhuangzi|publisher=Internet Encyclopedia of Philosophy|date=3 March 2016|url=http://www.iep.utm.edu/zhuangzi/|archive-url=https://web.archive.org/web/20160303175106/http://www.iep.utm.edu/zhuangzi/|archivedate=3 March 2016|url-status=live|access-date=5 March 2019|ref=harv}}
	* {{cite book|last=De George|first=Richard T.|editor=Ted Honderich|editor-link=Ted Honderich|title=The Oxford Companion to Philosophy|publisher=[[Oxford University Press]]|isbn= 9780199264797|year=2005|ref=harv}}
	* {{cite encyclopedia|last=Fiala|first=Andrew|title=Anarchism|encyclopedia=[[Stanford Encyclopedia of Philosophy]]|year=2017|url=https://plato.stanford.edu/entries/anarchism/|ref=harv|publisher=Metaphysics Research Lab, Stanford University}}
	* {{cite book|last1=McLean|first1=Iain|first2=Alistair|last2=McMillan|title=The Concise Oxford Dictionary of Politics|url=https://archive.org/details/oxfordconcisedic00iain|url-access=registration|year=2003|publisher=Oxford University Press|isbn=978-0-19-280276-7|ref=harv}}
	* {{cite web|title=Definition of Anarchism|work=Merriam-Webster|year=2019|url=https://www.merriam-webster.com/dictionary/anarchism|ref={{sfnref|Merriam-Webster|2019}}|access-date=28 February 2019}}
	* {{cite book|last=Miller|first=David|title=The Blackwell Encyclopaedia of Political Thought |url = https://books.google.com/books?id=NIZfQTd3nSMC|date=26 August 1991|publisher=Wiley|isbn=978-0-631-17944-3|ref=harv}}
	* {{cite book|last=Ostergaard|first=Geoffrey|year=2003|author-link=Geoffrey Ostergaard |title=The Blackwell Dictionary of Modern Social Thought|publisher=Blackwell Publishing|ref=harv}}
	* {{cite book|chapter=Anarchy|title=Oxford English Dictionary |edition=3rd |publisher=Oxford University Press|date= September 2005|ref={{harvid|Oxford English Dictionary|2005}} }} &lt;small&gt;Subscription or UK public library membership required.&lt;/small&gt;
	{{refend}}

	== Further reading ==
	* {{cite book |last=Barclay|first=Harold B. |author-link=Harold Barclay |title=People Without Government: An Anthropology of Anarchy |url = https://books.google.com/books?id=MrFHQgAACAAJ |year=1990|publisher=Kahn &amp; Averill|isbn=978-0-939306-09-1}}
	* {{cite book|last=Edmundson|first=William A. |title=Three Anarchical Fallacies: An Essay on Political Authority|url=https://books.google.com/books?id=q_gClKUbJyYC|date=2007|publisher=Cambridge University Press|isbn=978-0-521-03751-8}} Criticism of philosophical anarchism.
	* {{cite book |last=Harper|first=Clifford|authorlink=Clifford Harper|title=Anarchy: A Graphic Guide |url = https://books.google.com/books?id=W63aAAAAMAAJ |year=1987|publisher=Camden Press|isbn=978-0-948491-22-1}}
	* {{cite book |last=Le Guin|first=Ursula K. |author-link=Ursula K. Le Guin|title=The Dispossessed|date=2009|publisher=HarperCollins|title-link=The Dispossessed}}  Anarchistic popular fiction novel &lt;!-- Gifford 2019, p 580--&gt;
	* {{cite book |last=Kinna|first=Ruth|author-link=Ruth Kinna |title=Anarchism: A Beginners Guide |url = https://books.google.com/books?id=LLLaAAAAMAAJ |year=2005 |publisher=Oneworld |isbn=978-1-85168-370-3 }}
	* {{cite book |last=Sartwell|first=Crispin|author-link=Crispin Sartwell|title=Against the State: An Introduction to Anarchist Political Theory |publisher=SUNY Press|year=2008|isbn=978-0-7914-7447-1 |url = https://books.google.com/books?id=bk-aaMVGKO0C }}
	* {{cite book |last=Scott|first=James C. |author-link=James C. Scott |year=2012|title=Two Cheers for Anarchism: Six Easy Pieces on Autonomy, Dignity, and Meaningful Work and Play |title-link=Two Cheers for Anarchism|location=Princeton, New Jersey|publisher=Princeton University Press|isbn=978-0-691-15529-6}}
	* {{cite book |last=Wolff|first=Robert Paul|author-link=Robert Paul Wolff|title=In Defense of Anarchism|year=1998|publisher=University of California Press|isbn=978-0-520-21573-3|title-link=In Defense of Anarchism }} An argument for philosophical anarchism

	== External links ==
	{{sister project links|voy=no|n=no|v=no|b=Subject:Anarchism|s=Portal:Anarchism|d=Q6199|c=Category:Anarchism}}
	* [http://dwardmac.pitzer.edu/ Anarchy Archives]. [[Anarchy Archives]] is an online research center on the history and theory of anarchism
	&lt;!-- Attention! The external link portion of this article regularly grows far beyond manageable size. PLEASE only list an outside link if it applies to anarchism in general, is somewhat noteworthy, and has consensus on the talkpage. Links to sites which cover anarchist submovements will be routinely moved to subarticles to keep this article free of clutter. --&gt;
	{{prone to spam|date=November 2014}}
	{{Z148}}&lt;!-- {{no more links}}. Please be cautious adding more external links. Wikipedia is not a collection of links and should not be used for advertising. Excessive or inappropriate links will be removed. See [[Wikipedia:External links]] and [[Wikipedia:Spam]] for details. If there are already suitable links, propose additions or replacements on the article's talk page, or submit your link to the relevant category at Curlie (curlie.org) – and link there using {{curlie}}. --&gt;

	{{-}}
	{{anarchism}}
	{{anarchies}}
	{{libertarian socialism}}
	{{libertarianism}}
	{{philosophy topics}}
	{{political culture}}
	{{political ideologies}}
	{{social and political philosophy}}
	{{authority control}}

	[[Category:Anarchism| ]]
	[[Category:Anti-capitalism]]
	[[Category:Anti-fascism]]
	[[Category:Economic ideologies]]
	[[Category:Far-left politics]]
	[[Category:Libertarian socialism]]
	[[Category:Libertarianism]]
	[[Category:Political culture]]
	[[Category:Political movements]]
	[[Category:Political ideologies]]
	[[Category:Social theories]]</text>
		<sha1>bm226lfkmg6pktr3isb6b6znnwallfs</sha1>
	</revision>
</page>
`

var downloadContents = fmt.Sprintf(`<mediawiki xmlns="http://www.mediawiki.org/xml/export-0.10/" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.mediawiki.org/xml/export-0.10/ http://www.mediawiki.org/xml/export-0.10.xsd" version="0.10" xml:lang="en">
	%s
	%s
	%s
</mediawiki>
`, siteInfo, accessibleComputingXML, anarchismXML)
