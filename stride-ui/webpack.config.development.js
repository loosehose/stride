var webpack = require('webpack');
var path = require('path');

var config = {
    target: 'electron-renderer',
    externals: [
        {
            'electron-config': 'electron-config',
        },
    ],
    entry: [
        'webpack-hot-middleware/client?reload=true&path=http://localhost:9000/__webpack_hmr',
        './src/index',
    ],
    module: {
        rules: [
            {
                test: /\.jsx?$/,
                exclude: /(node_modules)/,
                use: {
                    loader: 'babel-loader',
                    options: {},
                },
            },
            {
                test: /\.css$/,
                use: [
                    'style-loader',
                    {
                        loader: 'css-loader',
                        options: {
                            importLoaders: 1,
                            modules: true
                        }
                    }
                ]
            },
            // Add loader for images
            {
                test: /\.(png|svg|jpg|jpeg|gif)$/i,
                type: 'asset/resource',
                generator: {
                  filename: 'images/[hash][ext][query]' // Outputs images to /dist/images
                }
            },
            // Add loader for SCSS files
            {
                test: /\.scss$/,
                use: [
                    'style-loader', // Injects CSS into the DOM via a <style> tag
                    'css-loader',   // Turns CSS into CommonJS
                    'sass-loader'   // Compiles Sass to CSS
                ],
            }
        ],
    },
    output: {
        path: path.resolve(__dirname, 'dist'),
        publicPath: 'http://localhost:9000/dist/',
        filename: 'bundle.js',
    },
    resolve: {
        extensions: ['.js', '.jsx'],
        alias: {
            utils: path.resolve(__dirname, 'src', 'js', 'utils.js'),
            modals: path.resolve(__dirname, 'src', 'components', 'Modals'),
        },
    },
    plugins: [new webpack.HotModuleReplacementPlugin()],
    node: {
        __dirname: false,
        __filename: false,
    },
    mode: 'development',
};
module.exports = config;
