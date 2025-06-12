import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  name: 'prettyPrintXml',
  standalone: true,
})
export class PrettyPrintXmlPipe implements PipeTransform {
  // source: https://stackoverflow.com/questions/376373/pretty-printing-xml-with-javascript
  transform(xmlString: string): string {
    let formatted = '',
      indent = '';
    let tab = '  ';
    xmlString.split(/>\s*</).forEach(function (node) {
      if (node.match(/^\/\w/)) indent = indent.substring(tab.length); // decrease indent by one 'tab'
      formatted += indent + '<' + node + '>\r\n';
      if (node.match(/^<?\w[^>]*[^\/]$/)) indent += tab; // increase indent
    });
    return formatted.substring(1, formatted.length - 3);
  }
}
