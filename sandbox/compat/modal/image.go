package modal

import (
	"context"
	"fmt"
	"strings"

	"github.com/digitalocean/godo/sandbox"
	"github.com/digitalocean/godo/sandbox/compat/internal/shared"
)

// Image mirrors modal.Image. A registry image becomes a DigitalOcean template
// built from that OCI image. The RUN, ENV and WORKDIR lines that
// DockerfileCommands adds are replayed in the sandbox after it starts.
type Image struct {
	ImageID string

	client   *Client
	base     string
	commands []string
	// problem is set when the image uses something that is not supported, such
	// as a private registry. Build and sandbox creation return it.
	problem error
}

// imageService implements ImageService.
type imageService struct {
	notImplementedImageService
	c *Client
}

func newImage(c *Client, base string, commands []string, problem error) *Image {
	img := &Image{client: c, base: base, commands: commands, problem: problem}
	img.ImageID = hashID("im", append([]string{base}, commands...)...)
	c.remember(img)
	return img
}

// FromRegistry mirrors ImageService.FromRegistry. A private registry secret is
// not supported.
func (s imageService) FromRegistry(tag string, params *ImageFromRegistryParams) *Image {
	var problem error
	if params != nil && params.Secret != nil {
		problem = optionNotImplemented("Images.FromRegistry", "Secret", "a private registry credentials API")
	}
	return newImage(s.c, tag, nil, problem)
}

// FromAwsEcr mirrors ImageService.FromAwsEcr. It is not supported: the image
// is returned, and Build and sandbox creation fail with the error.
func (s imageService) FromAwsEcr(tag string, secret *Secret, params *ImageFromAwsEcrParams) *Image {
	return newImage(s.c, tag, nil, notImplemented("Images.FromAwsEcr", "a private registry credentials API"))
}

// FromGcpArtifactRegistry mirrors ImageService.FromGcpArtifactRegistry. It is
// not supported, like FromAwsEcr.
func (s imageService) FromGcpArtifactRegistry(tag string, secret *Secret, params *ImageFromGcpArtifactRegistryParams) *Image {
	return newImage(s.c, tag, nil, notImplemented("Images.FromGcpArtifactRegistry", "a private registry credentials API"))
}

// FromID mirrors ImageService.FromID. It finds images made by this client.
func (s imageService) FromID(ctx context.Context, imageID string, params *ImageFromIDParams) (*Image, error) {
	s.c.mu.Lock()
	img := s.c.images[imageID]
	s.c.mu.Unlock()
	if img == nil {
		return nil, NotFoundError{Exception: fmt.Sprintf("image %s not found: only images made by this client can be found", imageID)}
	}
	return img, nil
}

// DockerfileCommands mirrors Image.DockerfileCommands. It returns a new image;
// only FROM, RUN, ENV and WORKDIR lines are supported. Secrets and GPU on the
// layer are not supported; ForceBuild has no effect because layers are not cached.
func (image *Image) DockerfileCommands(commands []string, params *ImageDockerfileCommandsParams) *Image {
	problem := image.problem
	cmds := append([]string{}, image.commands...)
	if params != nil {
		switch {
		case len(params.Secrets) > 0:
			problem = firstError(problem, optionNotImplemented("Image.DockerfileCommands", "Secrets", "an image build secrets API"))
		case params.GPU != "":
			problem = firstError(problem, optionNotImplemented("Image.DockerfileCommands", "GPU", "a GPU sandbox API"))
		}
		for _, k := range sortedKeys(params.Env) {
			cmds = append(cmds, fmt.Sprintf("ENV %s=%s", k, params.Env[k]))
		}
	}
	cmds = append(cmds, commands...)
	return newImage(image.client, image.base, cmds, problem)
}

// Build mirrors Image.Build: it builds the template for the image's base now.
func (image *Image) Build(ctx context.Context, app *App, params *ImageBuildParams) (*Image, error) {
	if _, err := image.plan(ctx); err != nil {
		return nil, err
	}
	return image, nil
}

// imagePlan is what a sandbox needs from an image.
type imagePlan struct {
	template string
	env      map[string]string
	workdir  string
	steps    []string // shell commands to run after the sandbox starts
}

func (image *Image) plan(ctx context.Context) (*imagePlan, error) {
	if image.problem != nil {
		return nil, image.problem
	}
	p := &imagePlan{env: map[string]string{}}
	for _, line := range image.commands {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		op, arg, _ := strings.Cut(line, " ")
		arg = strings.TrimSpace(arg)
		switch strings.ToUpper(op) {
		case "FROM":
		case "RUN":
			p.steps = append(p.steps, arg)
		case "ENV":
			kv, err := parseEnv(arg)
			if err != nil {
				return nil, err
			}
			for k, v := range kv {
				p.env[k] = v
			}
		case "WORKDIR":
			p.workdir = arg
			p.steps = append(p.steps, "mkdir -p "+shellQuote(arg))
		default:
			return nil, optionNotImplemented("Image.DockerfileCommands", strings.ToUpper(op), "an image build API (only FROM, RUN, ENV and WORKDIR are supported)")
		}
	}
	name, err := image.client.templateFor(ctx, image.base)
	if err != nil {
		return nil, err
	}
	p.template = name
	return p, nil
}

// templateFor finds or builds the template for an OCI reference.
func (c *Client) templateFor(ctx context.Context, ref string) (string, error) {
	c.mu.Lock()
	name, ok := c.templates[ref]
	c.mu.Unlock()
	if ok {
		return name, nil
	}
	name = shared.TemplateName("modal", ref)
	found := false
	for t, err := range c.do.ListTemplates(nil).All(ctx) {
		if err != nil {
			return "", err
		}
		if t.Name() == name {
			found = true
			if err := t.WaitUntilReady(ctx, nil); err != nil {
				return "", err
			}
			break
		}
	}
	if !found {
		t, err := c.do.CreateTemplate(ctx, &sandbox.TemplateParams{Name: name, SourceOCIRef: ref})
		if err != nil {
			return "", err
		}
		if err := t.WaitUntilReady(ctx, nil); err != nil {
			return "", err
		}
	}
	c.mu.Lock()
	c.templates[ref] = name
	c.mu.Unlock()
	return name, nil
}

func parseEnv(arg string) (map[string]string, error) {
	out := map[string]string{}
	fields := strings.Fields(arg)
	if len(fields) >= 2 && !strings.Contains(fields[0], "=") {
		out[fields[0]] = strings.Trim(strings.Join(fields[1:], " "), `"'`)
		return out, nil
	}
	for _, f := range fields {
		k, v, ok := strings.Cut(f, "=")
		if !ok {
			return nil, fmt.Errorf("modal: cannot parse ENV %q", arg)
		}
		out[k] = strings.Trim(v, `"'`)
	}
	return out, nil
}

func shellQuote(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

func firstError(a, b error) error {
	if a != nil {
		return a
	}
	return b
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}
