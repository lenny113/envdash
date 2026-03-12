# Git Hygiene for the project

This document outlines the rules for how we will use Git for the project.

It assumes the developer uses **Sublime Merge** for the project.
Sublime Merge can be downloaded here: [https://www.sublimemerge.com/download](https://www.sublimemerge.com/download)

This can be also installed by running(linux not tested):

#### Windows

```
winget install --id SublimeHQ.SublimeMerge
```

#### Linux

```
sudo snap install sublime-merge --classic
```


## Git commit workflow

This section functions as a tutorial and a standard for how we should add code to the project.

### Branch creation
If you want to push something first check out with 
```
git checkout -b name-of-branch
```
where the branch name is a sensible name.

After which you add the relevant files. 
**DO NOT USE GIT ADD . EVER** It is very easy to accidentally add API keys or irrelevant files to git when doing this. 
Add the specific files you are working on, if you are using sublime merge you can also specify which lines you wish to add to the commit.

### When pushing
Pull from main before pushing using 
```
git pull origin main
```
and resolve any merge conflicts locally.
**NEVER PUSH MERGE CONFLICTS TO MAIN**

### Merging
After pushing create a merge request to main, and add the other contributors as reviewers.
After which you can approve the merge after a review. Remember to check off delete sourc branch after merge.

## Branch policy

Main is the master branch. This is protected and is not to be pushed directly to.

Merge requests should ideally always be approved, however contributors are allowed to stage merges themselves.

## Commit message hygiene
each commit will follow this template:

```
# Commit message template
# scope(type): short header explaining the change

# one line explaining how this works today


# one line explaining why the commit is necessary


# one or more lines explaining the solution / tradeoffs
```
you can install this template, so that it opens by default in sublime merge by running the code below

### Install the commit template

You can configure the commit template automatically by running the setup script included in the repository:

#### Windows

```
bash playground/devutils/setup-git-template.sh
```

#### Linux

```
sh playground/devutils/setup-git-template.sh
```

This will configure Git to use the commit template for this repository.

If you prefer to configure it manually, follow the steps below.

Ensure the template file exists at:

```
playground/devutils/gitmessage.txt
```

Then configure Git to use this template by running the following command from the repository root:

```
git config commit.template playground/devutils/gitmessage.txt
```

This sets the commit template for the current repository only.

You can verify the configuration with:

```
git config --get commit.template
```

If configured correctly, this should output:

```
playground/devutils/gitmessage.txt
```

After this is set, the commit message template will automatically appear when creating commits in Git tools such as Sublime Merge or when running `git commit` from the terminal.
