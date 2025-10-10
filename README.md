# API Documentation

This branch contains the automatically generated API documentation for both `master` and `develop` branches.

## Structure

```
docs/
├── index.html          # Main landing page
├── master/             # Production API documentation
│   ├── index.html      # Swagger UI for master branch
│   ├── swagger.yaml    # OpenAPI specification
│   └── swagger.md      # Markdown documentation
└── develop/            # Development API documentation
    ├── index.html      # Swagger UI for develop branch
    ├── swagger.yaml    # OpenAPI specification
    └── swagger.md      # Markdown documentation
```

## Accessing Documentation

### GitHub Pages (Recommended)
If GitHub Pages is enabled for this repository:
- **Main Page**: `https://yourusername.github.io/yourrepo/`
- **Production API**: `https://yourusername.github.io/yourrepo/master/`
- **Development API**: `https://yourusername.github.io/yourrepo/develop/`

### Raw GitHub Files
- **Main Page**: `https://raw.githubusercontent.com/yourusername/yourrepo/docs/index.html`
- **Production API**: `https://raw.githubusercontent.com/yourusername/yourrepo/docs/master/index.html`
- **Development API**: `https://raw.githubusercontent.com/yourusername/yourrepo/docs/develop/index.html`

## How It Works

1. **Automatic Updates**: Documentation is automatically updated when code is pushed to `master` or `develop` branches
2. **Generated Content**: All documentation is generated from source code using Swagger/OpenAPI tools
3. **Branch Isolation**: Each branch maintains its own documentation version
4. **No Manual Maintenance**: This branch is managed entirely by GitHub Actions

## Workflow

The documentation is updated by the `update-docs-branch.yaml` GitHub Actions workflow:

1. Detects changes to API code in `master` or `develop`
2. Generates fresh Swagger documentation
3. Updates the corresponding directory in this branch
4. Commits and pushes changes automatically

## Files in This Branch

- `index.html` - Landing page with links to all documentation versions
- `master/` - Documentation for the production API
- `develop/` - Documentation for the development API
- `README.md` - This file

## Note

⚠️ **Do not manually edit files in this branch**. All content is automatically generated and will be overwritten by the GitHub Actions workflow.

For questions or issues with the documentation, please refer to the main repository or create an issue.