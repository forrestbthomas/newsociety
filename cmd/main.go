package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"sort"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

var (
	existingPosts []BlogPost = []BlogPost{}
	topics        []string   = []string{
		`Foundations of Justice and Equality
        Key Questions: What principles of justice should underpin the society? How can we ensure equality of opportunity, and what does it mean in practice? Philosophers like Martha Nussbaum and Amartya Sen's capabilities approach can provide a framework for addressing these questions, focusing on what individuals are able to do and be, rather than just what they have.
        Discussion Points: The role of the state in redistributing resources, ensuring access to healthcare, education, and housing, and addressing systemic inequalities.`,
		`The Role of Community and the Individual
        Key Questions: How do we balance individual autonomy with the needs and values of the community? Michael Sandel and Charles Taylor's work can guide discussions on the importance of communal values in individual development and the role of public discourse in shaping these values.
        Discussion Points: The importance of civic engagement, the promotion of public deliberation, and the cultivation of a shared sense of belonging and mutual responsibility.`,
		`Economic Structures and Social Welfare
        Key Questions: What economic models best support the ideals of the society? How can economic policies promote individual well-being and societal prosperity? Discussions could draw on the critiques and proposals of philosophers like Nancy Fraser, focusing on redistribution, recognition, and representation to address economic inequalities and social justice.
        Discussion Points: The feasibility and implications of different economic systems (e.g., capitalist, socialist, mixed economies), the role of social safety nets, and the importance of equitable access to resources.`,
		`Governance, Democracy, and Participation
        Key Questions: What forms of governance best facilitate active participation and ensure the society's democratic values? Jürgen Habermas's theory of communicative action can inform discussions on creating a political structure that encourages rational discourse and participatory democracy.
        Discussion Points: Mechanisms for ensuring transparency, accountability, and direct citizen involvement in decision-making processes; the role of digital technologies in enhancing democratic participation.`,
		`Cultural Identity, Diversity, and Integration
        Key Questions: How can the society respect and integrate diverse cultural identities while fostering a shared sense of community? Axel Honneth and Charles Taylor's work on recognition and the politics of recognition can guide discussions on the importance of acknowledging and valuing cultural differences as part of the social fabric.
        Discussion Points: Strategies for promoting multiculturalism and inclusion, addressing historical injustices, and the role of education in cultivating an appreciation for diversity.`,
		`Sustainability and Environmental Ethics
        Key Questions: How can the society ensure its development is sustainable and respects the natural environment? What ethical responsibilities do individuals and communities have towards the environment?
        Discussion Points: The importance of integrating environmental considerations into economic and social planning, promoting sustainable living practices, and addressing climate change. Philosophers could explore frameworks for balancing human needs with the preservation of the planet, drawing on concepts like the stewardship of nature and intergenerational justice.`,
		`Technology, Privacy, and Surveillance
        Key Questions: In an increasingly digital world, how does the society protect individual privacy while ensuring public safety and security? What role should technology play in the society, and how can it be used ethically?
        Discussion Points: The impact of technology on social interactions, work, and democracy, including the ethics of artificial intelligence, data protection laws, and the balance between surveillance for security and individual freedoms. Philosophers might discuss the implications of technological advancements on autonomy, competence, and interpersonal relationships, emphasizing the need for ethical guidelines that govern technology use.`,
		`Health, Well-being, and Access to Care
        Key Questions: What constitutes well-being in the society, and how can it be achieved for all citizens? How does the society ensure equitable access to healthcare and promote physical and mental health?
        Discussion Points: The definition of health as not merely the absence of disease but a state of complete physical, mental, and social well-being. Discussions could focus on the design of healthcare systems, the importance of preventative care, mental health services, and the ethical considerations surrounding end-of-life care. Philosophers could explore the balance between individual lifestyle choices and societal responsibilities in promoting health, drawing on concepts of social determinants of health and the right to healthcare.`,
		`Ethics and Governance of Artificial Intelligence (AI)
        Key Questions: How can society ensure that the development and deployment of AI technologies align with ethical principles and contribute positively to societal well-being? What governance structures are necessary to oversee AI research, development, and implementation to prevent harm and ensure accountability?   
        Discussion Points:
        Ethical AI Design and Use: Exploring frameworks for ethical AI that prioritize transparency, fairness, and privacy. This includes the development of AI systems that are explainable, non-discriminatory, and respect user consent, addressing concerns around bias, manipulation, and surveillance.
        AI and Employment: Addressing the impact of AI and automation on the workforce, including the displacement of jobs and the ethics of delegating tasks to AI. Philosophers could discuss ways to ensure that the benefits of AI are broadly shared across society, considering universal basic income or re-skilling programs as potential responses.
        AI Decision-Making: The implications of using AI in decision-making processes in areas like criminal justice, healthcare, and finance. This includes the ethical considerations of delegating life-impacting decisions to algorithms and the need for human oversight and appeal mechanisms.
        Global AI Governance: The necessity for international collaboration in creating standards and regulations for AI development and use, ensuring that AI technologies do not exacerbate global inequalities or undermine international security.
        AI and Human Identity: Exploring philosophical questions about the nature of consciousness, personhood, and the relationship between humans and intelligent machines. This includes discussions on the potential for AI to challenge or redefine concepts of creativity, empathy, and autonomy.`,
	}
	philosophers []string = []string{
		"Martha Nussbaum",
		"Amartya Sen",
		"Michael Sandel",
		"Charles Taylor",
		"Jürgen Habermas",
		"Axel Honneth",
		"Nancy Fraser",
		"Immanuel Kant",
		"John Dewey",
		"Hannah Arendt",
		"Michel Foucault",
		"Confucius",
		"Thomas Aquinas",
		"Simone de Beauvoir",
		"John Rawls",
		"Friedrich Nietzsche",
		"Martin Heidegger",
		"Noam Chomsky",
	}
	openAIKey string = ""
)

const model = openai.GPT4TurboPreview

// Define a struct to hold blog post data
type BlogPost struct {
	Title        string        `json:"Title"`
	Content      template.HTML `json:"Content"`
	CreationDate time.Time     `json:"CreationDate"`
}

// Define a struct for the page that includes a slice of BlogPost
type PageData struct {
	PageTitle string
	Posts     []BlogPost
}

func init() {
	openAIKey = getOpenAIKey()
	if openAIKey == "" {
		fmt.Println("no Open AI Key found")
		os.Exit(1)
	}
}

func getOpenAIKey() string {
	key, ok := os.LookupEnv("OPENAI_AICLASSICS_API_KEY2")
	if !ok {
		return ""
	}
	return key
}

func main() {
	client := openai.NewClient(openAIKey)
	tmpl := template.Must(template.New("blog").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.PageTitle}}</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            color: #333;
            margin: 0;
            padding: 0;
            background-color: #f4f4f4;
        }
        .container {
            max-width: 800px;
            margin: auto;
            background: #fff;
            padding: 20px;
        }
        header, footer {
            text-align: center;
            padding: 20px 0;
        }
        header h1 {
            margin: 0;
        }
        article {
            border-bottom: 1px solid #eaeaea;
            margin-bottom: 20px;
            padding-bottom: 20px;
        }
        h2 {
            color: #333;
        }
        p {
            line-height: 1.6;
        }
        footer p {
            margin: 0;
            color: #666;
        }
    </style>
</head>
<body>

<div class="container">
    <header>
        <h1>{{.PageTitle}}</h1>
    </header>

    {{range .Posts}}
    <article>
        {{.Content}}
        <h3>{{.CreationDate.Format "2006-01-02"}}</h3>
    </article>
    {{end}}

    <footer>
        <p>&copy; 2024 Enough is Enough</p>
    </footer>
</div>

</body>
</html>
`))

	existingPosts, err := GetExistingBlogPosts()
	if err != nil {
		fmt.Println(err)
		return
	}
	newPost, err := AskChatGPT(client)
	if err != nil {
		fmt.Println("Problem getting new Post", err)
		return
	}

	posts := append([]BlogPost{newPost}, existingPosts...)

	// Data for the blog
	data := PageData{
		PageTitle: "Enough Is Enough",
		Posts:     posts,
	}

	// Sort Posts in descending order by CreationDate
	sort.Slice(data.Posts, func(i, j int) bool {
		return data.Posts[i].CreationDate.After(data.Posts[j].CreationDate)
	})

	// Create a file to write the generated HTML
	file, err := os.Create("index.html")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Execute the template, writing the generated HTML to the file
	err = tmpl.Execute(file, data)
	if err != nil {
		panic(err)
	}

	println("Blog HTML generated successfully.")

}

func AskChatGPT(client *openai.Client) (BlogPost, error) {
	prompt := fmt.Sprintf(`
    Create a 500 to 2000 word blog post on one of the topics below (your choice), from the perspective of one of the philosophers below (your choice).  

    Topics
    %s
    
    Make sure to include quotes from at least 3 others from the list of the philosophers.
    Philosophers
    %s
    
    Your response should be in the following JSON format:
    {
        Title: "Title of Post as a string",
        Content: "Content of the post in HTML format - Structure the content as a blog post with proper tags, styles, etc, and assume the content starts in the middle of a pre-existing html page"
    }

    When sending the response, I want you to take on the role of an expert web developer who can style the blog as follows:
    - Modern minimalist
    - Grayscale color pallette
    - No Javascript or external libraries
    - In-line styles
    `,
		strings.Join(topics, "\n"), strings.Join(philosophers, "\n"))

	response, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: "json_object",
		},
	})
	if err != nil {
		return BlogPost{}, err
	}

	blogPostContent := template.HTML(response.Choices[0].Message.Content)
	blogPost := BlogPost{}
	err = json.Unmarshal([]byte(blogPostContent), &blogPost)
	if err != nil {
		return BlogPost{}, err
	}

	fileName := base64.RawStdEncoding.EncodeToString([]byte(blogPost.Title))
	err = os.WriteFile(fmt.Sprintf("content/%s.json", fileName), []byte(blogPostContent), 0644)
	if err != nil {
		return BlogPost{}, err
	}

	blogPost.CreationDate = time.Now()
	return blogPost, err
}

func GetExistingBlogPosts() (posts []BlogPost, err error) {
	files, err := os.ReadDir("content")
	if err != nil {
		return posts, err
	}
	for _, file := range files {
		p := BlogPost{}
		fileBytes, err := os.ReadFile(fmt.Sprintf("content/%s", file.Name()))
		if err != nil {
			return posts, err
		}
		err = json.Unmarshal(fileBytes, &p)
		if err != nil {
			return posts, err
		}
		posts = append(posts, p)
	}
	return posts, nil
}
