package utils

import (
	"net/http"
	"path/filepath"

	"github.com/apex/log"

	"github.com/pkg/errors"
	gitlab "github.com/xanzy/go-gitlab"
)

func is2xx(s int) bool {
	return s >= 200 && s <= 299
}

func checkResponse(res *gitlab.Response, err error, acceptable func(int) bool) error {
	if err != nil {
		return err
	}
	if !acceptable(res.StatusCode) {
		return errors.Errorf("unexpected status code: %v", res.StatusCode)
	}
	return nil
}

// GitlabApi is our wrapper tool around gitlab.Client
type GitlabApi struct {
	Client *gitlab.Client
}

// NewGitlabApi creates a new gitlab api and returns it
func NewGitlabApi(httpClient *http.Client, apiBaseURL string, privateToken string) *GitlabApi {
	optURL := gitlab.WithBaseURL(apiBaseURL)
	optHTTP := gitlab.WithHTTPClient(httpClient)
	client, err := gitlab.NewClient(privateToken, optURL, optHTTP)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	return &GitlabApi{
		Client: client,
	}
}

func (h *GitlabApi) EnumGroupProjects(group *gitlab.Group) (<-chan *gitlab.Project, <-chan error) {
	outc := make(chan *gitlab.Project, 10)
	errc := make(chan error)
	go func() {
		defer close(outc)
		defer close(errc)
		opts := &gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{
				// PerPage: 20,
				Page: 1,
			},
		}
		for {
			ps, resp, err := h.Client.Groups.ListGroupProjects(group.ID, opts)
			if err := checkResponse(resp, err, is2xx); err != nil {
				errc <- errors.Wrap(err, "listing group projects")
				return
			}
			for _, p := range ps {
				outc <- p
			}
			// Exit the loop when we've seen all pages.
			if resp.CurrentPage >= resp.TotalPages {
				break
			}
			// Update the page number to get the next page.
			opts.Page = resp.NextPage
		}
	}()
	return outc, errc
}

func (h *GitlabApi) EnumGroupsProjects(groups <-chan *gitlab.Group) (<-chan *gitlab.Project, <-chan error) {
	outc := make(chan *gitlab.Project, 10)
	errc := make(chan error)
	go func() {
		defer close(outc)
		defer close(errc)
		for group := range groups {
			pros, errs := h.EnumGroupProjects(group)
			select {
			case pro := <-pros:
				outc <- pro
			case err := <-errs:
				errc <- err
				break
			}
		}
	}()
	return outc, errc
}

func (h *GitlabApi) EnumGroups() (<-chan *gitlab.Group, <-chan error) {
	outc := make(chan *gitlab.Group, 10)
	errc := make(chan error)
	go func() {
		defer close(outc)
		defer close(errc)
		opts := &gitlab.ListGroupsOptions{
			ListOptions: gitlab.ListOptions{
				Page: 1,
			},
		}
		for {
			gs, resp, err := h.Client.Groups.ListGroups(opts)
			if err := checkResponse(resp, err, is2xx); err != nil {
				errc <- errors.Wrap(err, "listing groups")
				return
			}
			for _, g := range gs {
				outc <- g
			}
			if resp.CurrentPage >= resp.TotalPages {
				break
			}
			opts.Page = resp.NextPage
		}
	}()
	return outc, errc
}

func (h *GitlabApi) EnumAllGroupsProjects() (<-chan *gitlab.Project, <-chan error) {
	outc := make(chan *gitlab.Project, 10)
	errc := make(chan error)
	go func() {
		defer close(outc)
		defer close(errc)
		groups, gerrs := h.EnumGroups()
		pros, perrs := h.EnumGroupsProjects(groups)
		for {
			select {
			case pro := <-pros:
				outc <- pro
			case err := <-gerrs:
				errc <- err
				break
			case err := <-perrs:
				errc <- err
				break
			}
		}
	}()
	return outc, errc
}

func (h *GitlabApi) ListGroups(opts *gitlab.ListGroupsOptions) ([]*gitlab.Group, error) {
	all := make([]*gitlab.Group, 0)
	for {
		if opts.Page == 0 {
			opts.Page = 1
		}
		groups, resp, err := h.Client.Groups.ListGroups(opts)
		if err := checkResponse(resp, err, is2xx); err != nil {
			return nil, errors.Wrap(err, "listing groups")
		}
		all = append(all, groups...)
		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		opts.Page = resp.NextPage
	}
	return all, nil
}

func (h *GitlabApi) ListGroupProjects(group *gitlab.Group) ([]*gitlab.Project, error) {
	log.WithFields(log.Fields{"group": group.Name}).Debug("listing group")
	opts := &gitlab.ListGroupProjectsOptions{
		ListOptions: gitlab.ListOptions{
			Page: 1,
		},
	}
	all := make([]*gitlab.Project, 0)
	for {
		projects, resp, err := h.Client.Groups.ListGroupProjects(group.ID, opts)
		if err := checkResponse(resp, err, is2xx); err != nil {
			return nil, errors.Wrap(err, "listing group projects")
		}
		all = append(all, projects...)
		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		opts.Page = resp.NextPage
	}
	return all, nil
}

func (h *GitlabApi) ListAllGroupsProjects() ([]*gitlab.Project, error) {
	groups, err := h.ListGroups(&gitlab.ListGroupsOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "listing groups")
	}
	all := make([]*gitlab.Project, 0)
	for _, group := range groups {
		projects, err := h.ListGroupProjects(group)
		if err != nil {
			return nil, errors.Wrap(err, "listing groups projects")
		}
		all = append(all, projects...)
	}
	return all, nil
}

func (h *GitlabApi) ListProjectBranches(project *gitlab.Project) ([]*gitlab.Branch, error) {
	opts := &gitlab.ListBranchesOptions{
		ListOptions: gitlab.ListOptions{
			Page: 1,
		},
	}
	all := make([]*gitlab.Branch, 0)
	for {
		branches, resp, err := h.Client.Branches.ListBranches(project.ID, opts)
		if err := checkResponse(resp, err, is2xx); err != nil {
			return nil, errors.Wrap(err, "listing project branches")
		}
		all = append(all, branches...)
		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		opts.Page = resp.NextPage
	}
	return all, nil
}

func (h *GitlabApi) ListProjectTags(project *gitlab.Project) ([]*gitlab.Tag, error) {
	opts := &gitlab.ListTagsOptions{
		ListOptions: gitlab.ListOptions{
			Page: 1,
		},
	}
	all := make([]*gitlab.Tag, 0)
	for {
		tags, resp, err := h.Client.Tags.ListTags(project.ID, opts)
		if err := checkResponse(resp, err, is2xx); err != nil {
			return nil, errors.Wrap(err, "listing project tags")
		}
		all = append(all, tags...)
		// if res.TotalItems == 0 || res.CurrentPage == res.TotalPages {
		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		opts.Page = resp.NextPage
	}
	return all, nil
}

type Node struct {
	Path     string
	TreeNode *gitlab.TreeNode
}

func (h *GitlabApi) ListTree(project *gitlab.Project, ref string, path string) ([]*Node, error) {
	log.WithFields(log.Fields{"group": project.Namespace.Name, "project": project.Name, "ref": ref, "path": path}).Debug("listing tree")
	opts := &gitlab.ListTreeOptions{
		Ref:  &ref,
		Path: &path,
	}
	nodes, res, err := h.Client.Repositories.ListTree(project.ID, opts)
	if err := checkResponse(res, err, is2xx); err != nil {
		return nil, errors.Wrap(err, "listing tree nodes")
	}
	all := make([]*Node, 0)
	for _, node := range nodes {
		nodePath := filepath.Join(path, node.Name)
		switch node.Type {
		case "tree":
			nodeChildren, err := h.ListTree(project, ref, nodePath)
			if err != nil {
				return nil, errors.Wrap(err, "listing tree node children")
			}
			all = append(all, nodeChildren...)
		case "blob":
			all = append(all, &Node{Path: nodePath, TreeNode: node})
		default:
			log.WithFields(log.Fields{"type": node.Type}).Errorf("unknown node type")
		}
	}
	return all, nil
}

func (h *GitlabApi) GetUser(username string) (*gitlab.User, error) {
	users, res, err := h.Client.Users.ListUsers(&gitlab.ListUsersOptions{
		Search: &username,
	})
	if err := checkResponse(res, err, is2xx); err != nil {
		return nil, errors.Wrap(err, "listing users")
	}
	if len(users) > 1 {
		return nil, errors.Errorf("multiple users found for username '%s'", username)
	}
	if len(users) == 0 {
		return nil, errors.Errorf("no users found for username '%s'", username)
	}
	return users[0], nil
}

func (h *GitlabApi) ReplaceOwners(project *gitlab.Project, owners ...*gitlab.User) error {
	err := h.AddMembers(project, gitlab.AccessLevel(gitlab.OwnerPermission), owners...)
	if err != nil {
		return errors.Wrap(err, "adding project owners")
	}
	opts := &gitlab.ListProjectMembersOptions{
		ListOptions: gitlab.ListOptions{
			Page: 1,
		},
	}
	for {
		members, resp, err := h.Client.ProjectMembers.ListProjectMembers(project.ID, opts)
		if err := checkResponse(resp, err, is2xx); err != nil {
			return errors.Wrap(err, "listing project members")
		}
		for _, member := range members {
			if member.AccessLevel != gitlab.OwnerPermission {
				continue
			}
			remove := true
			for _, owner := range owners {
				if member.ID == owner.ID {
					remove = false
					break
				}
			}
			if remove {
				resp, err := h.Client.ProjectMembers.DeleteProjectMember(project.ID, member.ID)
				if err := checkResponse(resp, err, is2xx); err != nil {
					return errors.Wrap(err, "deleting project member")
				}
			}
		}
		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		opts.Page = resp.NextPage
	}
	return nil
}

func (h *GitlabApi) AddMembers(project *gitlab.Project, perm *gitlab.AccessLevelValue, members ...*gitlab.User) error {
	for _, member := range members {
		_, res, err := h.Client.ProjectMembers.AddProjectMember(project.ID, &gitlab.AddProjectMemberOptions{
			AccessLevel: perm,
			UserID:      &member.ID,
		})
		if err := checkResponse(res, err, is2xx); err != nil {
			return errors.Wrap(err, "adding project member")
		}
	}
	return nil
}

func (h *GitlabApi) CreateProjects(gitlabOwners []string, gitlabGroup string, gitlabProjects []string) error {
	owners := make([]*gitlab.User, 0)
	for _, user := range gitlabOwners {
		owner, err := h.GetUser(user)
		if err != nil {
			return errors.Wrap(err, "getting user")
		}
		owners = append(owners, owner)
	}
	gs, err := h.ListGroups(&gitlab.ListGroupsOptions{
		Search: &gitlabGroup,
	})
	if err != nil {
		return errors.Wrap(err, "listing groups")
	}
	groupFound := false
	var g *gitlab.Group
	for _, g = range gs {
		log.Infof("checking group %s", g.Name)
		if g.Name == gitlabGroup {
			groupFound = true
			log.Infof("using group %q (%d)", g.Name, g.ID)
			break
		}
	}
	if !groupFound {
		log.Infof("creating group %q", gitlabGroup)
		var r *gitlab.Response
		g, r, err = h.Client.Groups.CreateGroup(&gitlab.CreateGroupOptions{
			Path:       &gitlabGroup,
			Name:       &gitlabGroup,
			Visibility: gitlab.Visibility(gitlab.PublicVisibility),
		})
		if err != nil {
			return errors.Wrapf(err, "creating group %q", gitlabGroup)
		}
		if r.StatusCode/100 != 2 {
			return errors.Errorf("unexpected gitlab response %v creating group %q", r.StatusCode, gitlabGroup)
		}
		log.Infof("created group %q (%d)", gitlabGroup, g.ID)
	}
	projects, err := h.ListGroupProjects(g)
	if err != nil {
		return errors.Wrap(err, "listing group projects")
	}
	for _, name := range gitlabProjects {
		found := false
		for _, project := range projects {
			if project.Name == name {
				found = true
				log.Infof("found existing project %q in group %q (%d)", name, g.Name, g.ID)
				break
			}
		}
		if !(found) {
			p, r, err := h.Client.Projects.CreateProject(&gitlab.CreateProjectOptions{
				Name:        &name,
				NamespaceID: &g.ID,
				Visibility:  gitlab.Visibility(gitlab.PublicVisibility),
			})
			if err != nil {
				return errors.Wrapf(err, "creating project %q in group %q", name, gitlabGroup)
			}
			if r.StatusCode/100 != 2 {
				return errors.Errorf("unexpected gitlab response %v creating project %q in group %q", r.StatusCode, name, gitlabGroup)
			}
			log.Infof("created project %q (%d) in group %q (%d)", name, p.ID, gitlabGroup, g.ID)
			level := gitlab.MasterPermissions
			for _, owner := range owners {
				_, r, err := h.Client.ProjectMembers.AddProjectMember(p.ID, &gitlab.AddProjectMemberOptions{
					AccessLevel: &level,
					UserID:      &(owner.ID),
				})
				if err != nil {
					return errors.Wrapf(err, "adding project owner %q on project %q", name, gitlabGroup)
				}
				if r.StatusCode/100 != 2 {
					return errors.Errorf("unexpected gitlab response %v creating project %q in group %q", r.StatusCode, name, gitlabGroup)
				}
				log.Infof("added user %q (%d) as owner of project %q (%d) in group %q (%d)", owner.Username, owner.ID, p.Name, p.ID, g.Name, g.ID)
			}
		}
	}
	return nil
}

func (h *GitlabApi) EnableDeployKey(deployKey *gitlab.DeployKey, idProject int) error {
	_, res, err := h.Client.DeployKeys.AddDeployKey(idProject, &gitlab.AddDeployKeyOptions{
		Key:     &deployKey.Key,
		Title:   &deployKey.Title,
		CanPush: deployKey.CanPush,
	})
	if err := checkResponse(res, err, is2xx); err != nil {
		return errors.Wrapf(err, "adding deploy key (%d %s) in project %d", deployKey.ID, deployKey.Title, idProject)
	}
	return nil
}

func (h *GitlabApi) DisabledDeployKey(deployKey *gitlab.DeployKey, idProject int) error {
	res, err := h.Client.DeployKeys.DeleteDeployKey(idProject, deployKey.ID)
	if err := checkResponse(res, err, is2xx); err != nil {
		return errors.Wrapf(err, "deleting deploy key (%d %s) in project %d", deployKey.ID, deployKey.Title, idProject)
	}
	return nil
}

func (h *GitlabApi) GetDeployKey(idProject, idDeployKey int) (*gitlab.DeployKey, error) {
	deployKeyFind, resp, err := h.Client.DeployKeys.GetDeployKey(idProject, idDeployKey)
	if err := checkResponse(resp, err, is2xx); err != nil {
		return nil, errors.Wrapf(err, "find deployKey %d in project %d", idDeployKey, idProject)
	}
	return deployKeyFind, nil
}

func (h *GitlabApi) ListDeployKeys(idProject int) ([]*gitlab.DeployKey, error) {
	opts := &gitlab.ListProjectDeployKeysOptions{
		Page: 1,
	}
	all := make([]*gitlab.DeployKey, 0)
	for {
		keys, resp, err := h.Client.DeployKeys.ListProjectDeployKeys(idProject, opts)
		if err := checkResponse(resp, err, is2xx); err != nil {
			return nil, errors.Wrapf(err, "list project deploy keys")
		}
		all = append(all, keys...)
		if resp.CurrentPage >= resp.TotalPages {
			break
		}
		opts.Page = resp.NextPage
	}
	return all, nil
}
