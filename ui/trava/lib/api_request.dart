import 'dart:convert';

import 'package:http/http.dart' as http;

class ApiRequest {
  
  Future<String> getServerInfo() async {
    final response = await http.get(Uri.parse('localhost:8080/api/v1/server_info'));
    if (response.statusCode == 200) {
      var json = jsonDecode(response.body) as Map<String, dynamic>;
      return json['server']['name'] as String;
    } else {
      return "Trava TTT";
    }
  }

}
