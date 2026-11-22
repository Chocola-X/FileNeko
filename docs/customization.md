# Customization

**English** | [简体中文](customization_CN.md)

You can customize the styles, branding and icons of your FileNeko instance in order to give it a personal touch.

## Custom Branding

You can customize FileNeko to use your own branding. This includes the following:

- **Name**: the name of the instance that shows up on the tab title, login pages, and some other places.
- **Disable External Links**: disables all external links, except to the documentation.
- **Hide Version**: hides the software version on every web page, including public shares. The version is also omitted from the page's bootstrap data. Saving a change to this option reloads the current page; other open tabs apply it when refreshed. It does not change the CLI version command.
- **Disable Used Percentage**: disables the disk usage information on the sidebar.
- **Branding Folder**: directory which can contain two items:
  - `custom.css`, containing a global stylesheet to apply to all users.
  - `img`, a directory which can replace all the [default logotypes](https://github.com/filebrowser/filebrowser/tree/master/frontend/public/img) from the application.

This can be configured by the administrator user, under **Settings → Global Settings**. You can also update the configuration directly using the [CLI](cli/fileneko-config-set.md):

```sh
fileneko config set --branding.name "My Name" \
  --branding.files "/abs/path/to/my/dir" \
  --branding.disableExternal
```

Use `fileneko config set --branding.hideVersion` to hide the web version, or
`fileneko config set --branding.hideVersion=false` to show it again.

> [!NOTE] 
>
> If you are using Docker, you need to mount a volume with the `branding` directory in order for it to be accessible from within the container.

### Custom Icons

To replace the default logotype and favicons, you need to create an `img` directory under the branding directory. The structure of this directory must mimic the one from the [default logotypes](https://github.com/filebrowser/filebrowser/tree/master/frontend/public/img):

```
img/
  logo.svg
  icons/
    favicon.ico
    favicon.svg
    (...)
```

Note that there are different versions of the same favicon in multiple sizes. To replace all of them, you need to add versions for all of them. You can use the [Real Favicon Generator](https://realfavicongenerator.net/) to generate these for you from your base image. 

> [!NOTE]
>
> The icons are cached by the browser, so you may not see your changes immediately. You can address this by clearing your browser's cache.

## Theme Customization in This Fork

The built-in theme relies on standard CSS specificity and no longer uses `!important`. The overall appearance retains the original light theme styling, while administrators can customize the look and feel for all users via a standalone `custom.css` file within the branding directory, without modifying the frontend source code.

You can start from the [custom.css example](personalized-examples/custom.css) and remove unnecessary rules as needed. For a complete configuration featuring custom wallpapers, logos, fonts, and 90% translucent white cards, see the [Personalized Example](personalized-examples/README.md). Example directory structure:

```text
branding/
  custom.css
  img/
    logo.svg
    background.webp
    login.webp
    fonts/
      XiaolaiMono-Regular.woff2
```

`img/logo.svg` is used for both the login page and the top-left application logo after authentication. Using relative paths like `url("./img/background.webp")` in CSS ensures assets load correctly even when a subpath base URL such as `/library` is configured. Custom fonts should be hosted locally and defined using `@font-face`, while preserving system font fallbacks in `--font-ui`; never override the Material Icons icon font. Legacy themes relying on component hash attributes, outdated DOM hierarchies, or rigid layouts must be adapted and cannot simply be copied verbatim.

Prefer using the variables and stable classes documented in the [CSS Theme API v1](theme-api.md). Unset variables retain their default styling. The Theme API enumerates the names covered by stability and compatibility commitments; standard internal selectors can still be overridden when necessary.

Cards have an outer background and an inner decorative surface; when configuring transparency effects, adjust both layers together. Do not apply `opacity` to the entire card container, as that will make text, images, and buttons translucent as well. Palette variables such as `--card-main` are scoped to individual file cards; any `color-mix()` referencing these variables should be placed in card-level rules rather than evaluated globally on `:root`.

Finer adjustments can be made using standard selectors in `custom.css`. Because it is loaded after the built-in stylesheets, custom rules take precedence at equal specificity; if encountering more specific default selectors, prefer adjusting the designated theme variables. Numeric input adaptation rules are scoped to `#app`, and internal library class names are not part of the public Theme API.

Modifying branding assets does not require recompiling the application. However, browsers may cache assets for up to a day, so disable the cache or hard-refresh during testing. After initially configuring or changing the branding directory, reload the page. Before modifying database configurations via the CLI, ensure that any running service using that database has been stopped.

### Built-in Style Maintenance and Verification

CSS cascade layers are ordered as `vendor → base → components` and declared at the very beginning of the HTML document to prevent Vite's component CSS output from disrupting layer precedence. Third-party library styles reside in `vendor`, legacy layouts and base widgets in `base`, and Vue component styles in `components`. Built-in theme rules remain unlayered to correctly override runtime-injected third-party styles; the branding `custom.css` is loaded last and should likewise remain unlayered.

When adding new styles, modify the specific rule or theme variable responsible for the element rather than appending `!important`. The file list column widths are governed by CSS Grid rather than JavaScript stylesheet manipulation. `pnpm run test` verifies that no forced specificity (`!important`) exists in the project's own CSS, Vue components, or inline styles, and checks the placement of cascade layer declarations in HTML. Internal styles within third-party dependencies are outside this scope.

The visual regression scripts use temporary databases and files without connecting to existing deployments. They require Chromium, FFmpeg, and Python's `websocket-client` and `Pillow`:

```sh
python3 frontend/scripts/theme-regression.py /absolute/path/to/old-filebrowser /tmp/theme-before
python3 frontend/scripts/theme-regression.py /absolute/path/to/new-filebrowser /tmp/theme-after
python3 frontend/scripts/compare-theme.py /tmp/theme-before /tmp/theme-after
```

The regression suite covers desktop and mobile views for login, settings, numeric controls, six types of modal dialogs, search results, three file listing views, multi-selection operations, the sidebar, image/audio/video/PDF/EPUB/CSV previews, text editing, and Markdown preview. Prior to capturing screenshots, timestamps, capacity labels, and progress bars are frozen, hover states are cleared, and QR code canvases whose contents change with ephemeral share links are cleared (while preserving QR canvas dimensions and bounding boxes for layout comparison). The comparison report includes pixel diffs and computed style deltas; it complements but does not replace manual verification of newly added interactions.
