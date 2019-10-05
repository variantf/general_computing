const webpack = require('webpack');
const path = require('path');
const HtmlWebpackPlugin = require('html-webpack-plugin');
const ExtractTextPlugin = require('extract-text-webpack-plugin');

const __PROD__ = process.env.NODE_ENV === 'production';
const __DEV__ = !__PROD__;
const define = {
  __DEV__: JSON.stringify(__DEV__),
  __PROD__: JSON.stringify(__PROD__),
  // 是否内嵌调试工具
  __DEVTOOLS__: JSON.stringify(__DEV__),
  // React等库通过process.env.NODE_ENV变量判断是否运行于生产环境。
  'process.env': {
    NODE_ENV: JSON.stringify(__PROD__ ? 'production' : 'development')
  }
};

module.exports = {
  entry: [
    'whatwg-fetch',
    'normalize.css',
    path.join(__dirname, 'web/index.jsx')
  ],
  output: {
    path: path.join(__dirname, 'www'),
    filename: 'bundle.js',
    publicPath: ''
  },
  module: {
    loaders: [
      {
        test: /\.jsx?$/i,
        exclude: /node_modules/,
        loader: 'babel',
        query: {
          'passPerPreset': true,
          'presets': [
            {'plugins': ['transform-runtime']},
            {
              'passPerPreset': false,
              'presets': ['es2015', 'stage-0', 'react']
            }
          ]
        }
      }, {
        test: /\.(jpe?g|png|gif|svg)$/i,
        loader: 'file'
      }, {
        test: /\.css$/i,
        include: path.join(__dirname, 'web'),
        loader: ExtractTextPlugin.extract('style', 'css?modules&localIdentName=[local]___[hash:base64:5]')
      }, {
        test: /\.css$/i,
        exclude: path.join(__dirname, 'web'),
        loader: ExtractTextPlugin.extract('style', 'css')
      }
    ]
  },
  resolve: {
    extensions: ['', '.js', '.jsx']
  },
  devServer: {
	  compress: true,
	  public: 'sh-router.node.variantf.me',
  },
  plugins: [
    // 定义JavaScript中常量的值
    new webpack.DefinePlugin(define),
    // 生成一个index.html
    new HtmlWebpackPlugin({
      minify: {},
      template: path.join(__dirname, 'web/index.html')
    }),
    // 将CSS单独输出，提升加载速度
    new ExtractTextPlugin('style.css', {allChunks: true})
  ].concat(__PROD__? [
    // 优化并混淆JavaScript代码
    new webpack.optimize.UglifyJsPlugin({
      output: {comments: false},
      compress: {warnings: false},
      sourceMap: false
    })
  ]: [])
};
