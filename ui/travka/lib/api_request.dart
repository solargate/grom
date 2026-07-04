import 'dart:convert';

import 'package:http/http.dart' as http;

class ApiRequest {

  Future<String> getServerInfo() async {
    final response = await http.get(Uri.parse('/api/v1/server_info'));
    if (response.statusCode == 200) {
      var json = jsonDecode(response.body) as Map<String, dynamic>;
      return json['name'] as String;
    } else {
      return "Travka Home";
    }
  }

}
