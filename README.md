# API Documentation

This branch contains the automatically generated API documentation for both `master` and `develop` branches.

## Structure

```
docs/
├── index.html          # Main landing page
├── master/             # Production API documentation
│   ├── index.html      # Swagger UI for master branch
│   ├── swagger.yaml    # OpenAPI specification
└── develop/            # Development API documentation
    ├── index.html      # Swagger UI for develop branch
    ├── swagger.yaml    # OpenAPI specification
```

## Accessing Documentation

### GitHub Pages
If GitHub Pages is enabled for this repository:
- [Main Page](https://mini-maxit.github.io/backend/)
- [Production API](https://mini-maxit.github.io/backend/master/)
- [Development API](https://mini-maxit.github.io/backend/develop/)

## How It Works

1. **Automatic Updates**: Documentation is automatically updated when code is pushed to `master` or `develop` branches
2. **Generated Content**: All documentation is generated from source code using Swagger/OpenAPI tools
3. **Branch Isolation**: Each branch maintains its own documentation version
4. **No Manual Maintenance**: This branch is managed entirely by GitHub Actions

## Note

⚠️ **Do not manually edit files in this branch**. All content is automatically generated and will be overwritten by the GitHub Actions workflow.

For questions or issues with the documentation, please refer to the main repository or create an issue.
