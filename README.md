# WellWish Corporate Decision Engine

Wellwish is a platform similar to Kubernetes.

It allows you to run applets within your company that support decision-making.

## Design

The design is the result of three years of extensive research from 2020-2023 by Schmied Enterprises.

Productivity is the measure of the domestic product achieved per employee.
It has not improved much in decades, once data centers became mainstream.
The goal was to address the problem of productivity, with a simple, open benchmark solution.

The main concept behind WellWish is that is extremely supportable.
The marginal labor cost of running a cluster per node does not increase by the cluster size.
It requires no-dev, no-ops, no-os eventually. The concept is also known as the Personal Cloud.

There are two ways to achieve this.
One is that it uses Englang, plain words to describe code and data.
The second is that the entire state can easily be retrieved, stored, analyzed and ported.
Even an accountant can read the bare metal data files.

Kubernetes can do the same with pods that implement it.
The main difference is the mesh structure.
Kubernetes has a master, while Wellwish distributes even administrator requests across a unified cluster called the office.

Nodes of the same type scale better and cheaper.
Resource load differences between microservices are set within each node independently vs. the entire cluster.
Therefore, each node has a stateful container and many stateless burst containers that pick up requests and restart.
This structure also matches the design of the major cloud providers.

There are no roles. Each room in the office is a data or burst code that runs.
Each of them has a unique private key that you can use to knock-knock and use that micro-service.

The main business cost is low support costs.
Each structure is allowed to be twice as large as a classical binary or json.
This allows us to use Englang, help users, accountants, debuggers, etc.
Doubling the buffer size still scales well, as long as the office is scalable.

While it is designed to be scalable, we suggest using a cluster size of two nodes for optimal reliability.
A single node does not ensure scaling when it is needed.
One thousand nodes may bring in node errors with lost shards, where some customers run on older versions.
Two nodes ensure that any node errors surface fast, while they also ensure that scaling works,
and it is easy to add a third node when needed.

## Who is it for?

Wellwish targets a specific user base.

Creative Commons open source is suitable for biotech, healthcare, robotics research and development businesses, who are patent holders themselves.
Some copyright licenses other than Creative Commons may pose a patent risk for these companies.

Also, we target organizations that are low on devops resources.
The final goal is that professionals can use it, who can use tools like Microsoft Access, or Excel.

Please consult with a professional of your local jurisdiction.

## Getting started

To make it easy for you to get started with GitLab, here's a list of recommended next steps.

Already a pro? Just edit this README.md and make it your own. Want to make it easy? [Use the template at the bottom](#editing-this-readme)!

## Add your files

- [ ] [Create](https://docs.gitlab.com/ee/user/project/repository/web_editor.html#create-a-file) or [upload](https://docs.gitlab.com/ee/user/project/repository/web_editor.html#upload-a-file) files
- [ ] [Add files using the command line](https://docs.gitlab.com/ee/gitlab-basics/add-file.html#add-a-file-using-the-command-line) or push an existing Git repository with the following command:

```
cd existing_repo
git remote add origin https://gitlab.com/eper.io/wellwish-corporate-decision-engine.git
git branch -M main
git push -uf origin main
```

## Integrate with your tools

- [ ] [Set up project integrations](https://gitlab.com/eper.io/wellwish-corporate-decision-engine/-/settings/integrations)

## Collaborate with your team

- [ ] [Invite team members and collaborators](https://docs.gitlab.com/ee/user/project/members/)
- [ ] [Create a new merge request](https://docs.gitlab.com/ee/user/project/merge_requests/creating_merge_requests.html)
- [ ] [Automatically close issues from merge requests](https://docs.gitlab.com/ee/user/project/issues/managing_issues.html#closing-issues-automatically)
- [ ] [Enable merge request approvals](https://docs.gitlab.com/ee/user/project/merge_requests/approvals/)
- [ ] [Automatically merge when pipeline succeeds](https://docs.gitlab.com/ee/user/project/merge_requests/merge_when_pipeline_succeeds.html)

## Test and Deploy

Use the built-in continuous integration in GitLab.

- [ ] [Get started with GitLab CI/CD](https://docs.gitlab.com/ee/ci/quick_start/index.html)
- [ ] [Analyze your code for known vulnerabilities with Static Application Security Testing(SAST)](https://docs.gitlab.com/ee/user/application_security/sast/)
- [ ] [Deploy to Kubernetes, Amazon EC2, or Amazon ECS using Auto Deploy](https://docs.gitlab.com/ee/topics/autodevops/requirements.html)
- [ ] [Use pull-based deployments for improved Kubernetes management](https://docs.gitlab.com/ee/user/clusters/agent/)
- [ ] [Set up protected environments](https://docs.gitlab.com/ee/ci/environments/protected_environments.html)

***

# Editing this README

When you're ready to make this README your own, just edit this file and use the handy template below (or feel free to structure it however you want - this is just a starting point!). Thank you to [makeareadme.com](https://www.makeareadme.com/) for this template.

## Suggestions for a good README
Every project is different, so consider which of these sections apply to yours. The sections used in the template are suggestions for most open source projects. Also keep in mind that while a README can be too long and detailed, too long is better than too short. If you think your README is too long, consider utilizing another form of documentation rather than cutting out information.

## Name
Choose a self-explaining name for your project.

## Description
Let people know what your project can do specifically. Provide context and add a link to any reference visitors might be unfamiliar with. A list of Features or a Background subsection can also be added here. If there are alternatives to your project, this is a good place to list differentiating factors.

## Badges
On some READMEs, you may see small images that convey metadata, such as whether or not all the tests are passing for the project. You can use Shields to add some to your README. Many services also have instructions for adding a badge.

## Visuals
Depending on what you are making, it can be a good idea to include screenshots or even a video (you'll frequently see GIFs rather than actual videos). Tools like ttygif can help, but check out Asciinema for a more sophisticated method.

## Installation
Within a particular ecosystem, there may be a common way of installing things, such as using Yarn, NuGet, or Homebrew. However, consider the possibility that whoever is reading your README is a novice and would like more guidance. Listing specific steps helps remove ambiguity and gets people to using your project as quickly as possible. If it only runs in a specific context like a particular programming language version or operating system or has dependencies that have to be installed manually, also add a Requirements subsection.

## Usage
Use examples liberally, and show the expected output if you can. It's helpful to have inline the smallest example of usage that you can demonstrate, while providing links to more sophisticated examples if they are too long to reasonably include in the README.

## Support
Tell people where they can go to for help. It can be any combination of an issue tracker, a chat room, an email address, etc.

## Roadmap
If you have ideas for releases in the future, it is a good idea to list them in the README.

## Contributing
State if you are open to contributions and what your requirements are for accepting them.

For people who want to make changes to your project, it's helpful to have some documentation on how to get started. Perhaps there is a script that they should run or some environment variables that they need to set. Make these steps explicit. These instructions could also be useful to your future self.

You can also document commands to lint the code or run tests. These steps help to ensure high code quality and reduce the likelihood that the changes inadvertently break something. Having instructions for running tests is especially helpful if it requires external setup, such as starting a Selenium server for testing in a browser.

## Authors and acknowledgment
Show your appreciation to those who have contributed to the project.

## License
For open source projects, say how it is licensed.

## Project status
If you have run out of energy or time for your project, put a note at the top of the README saying that development has slowed down or stopped completely. Someone may choose to fork your project or volunteer to step in as a maintainer or owner, allowing your project to keep going. You can also make an explicit request for maintainers.
